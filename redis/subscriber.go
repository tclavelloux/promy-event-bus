package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/streams"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/semaphore"
)

// Subscriber implements EventSubscriber for Redis Streams.
type Subscriber struct {
	client *redis.Client
	config eventbus.Config
}

// NewSubscriber creates a new Redis subscriber.
func NewSubscriber(config eventbus.Config) (*Subscriber, error) {
	opts, err := redis.ParseURL(config.Redis.DSN)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis DSN: %w", err)
	}

	// Apply configuration
	if config.Redis.PoolSize > 0 {
		opts.PoolSize = config.Redis.PoolSize
	}
	if config.Redis.MaxRetries > 0 {
		opts.MaxRetries = config.Redis.MaxRetries
	}
	if config.Redis.MinRetryBackoff > 0 {
		opts.MinRetryBackoff = config.Redis.MinRetryBackoff
	}
	if config.Redis.MaxRetryBackoff > 0 {
		opts.MaxRetryBackoff = config.Redis.MaxRetryBackoff
	}
	if config.Redis.DialTimeout > 0 {
		opts.DialTimeout = config.Redis.DialTimeout
	}
	if config.Redis.ReadTimeout > 0 {
		opts.ReadTimeout = config.Redis.ReadTimeout
	}
	if config.Redis.WriteTimeout > 0 {
		opts.WriteTimeout = config.Redis.WriteTimeout
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Subscriber{
		client: client,
		config: config,
	}, nil
}

// Subscribe starts consuming events from Redis Streams.
//
//nolint:cyclop // Event loop functions naturally have higher complexity
func (s *Subscriber) Subscribe(ctx context.Context, subConfig eventbus.SubscriptionConfig) error {
	// Set defaults
	if subConfig.MaxConcurrency <= 0 {
		subConfig.MaxConcurrency = 1
	}
	if subConfig.BatchSize <= 0 {
		subConfig.BatchSize = 1
	}
	if subConfig.BlockDuration <= 0 {
		subConfig.BlockDuration = 1 * time.Second
	}

	// Create consumer group if it doesn't exist
	err := s.client.XGroupCreateMkStream(ctx, subConfig.Stream, subConfig.ConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("%w: failed to create consumer group: %w", eventbus.ErrSubscriptionFailed, err)
	}

	// Semaphore for concurrency control
	sem := semaphore.NewWeighted(int64(subConfig.MaxConcurrency))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read events from stream
			streams, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    subConfig.ConsumerGroup,
				Consumer: subConfig.ConsumerID,
				Streams:  []string{subConfig.Stream, ">"},
				Count:    int64(subConfig.BatchSize),
				Block:    subConfig.BlockDuration,
			}).Result()

			if err != nil {
				if errors.Is(err, redis.Nil) {
					continue // No messages available
				}

				return fmt.Errorf("%w: failed to read from stream: %w", eventbus.ErrSubscriptionFailed, err)
			}

			// Process messages concurrently
			var wg sync.WaitGroup
			for _, stream := range streams {
				for _, message := range stream.Messages {
					if err := sem.Acquire(ctx, 1); err != nil {
						return err
					}

					wg.Add(1)
					go func(msg redis.XMessage) {
						defer wg.Done()
						defer sem.Release(1)

						s.processMessage(ctx, subConfig, msg)
					}(message)
				}
			}

			wg.Wait()
		}
	}
}

// errMissingMetadata is reported when the metadata field is absent or not a string.
var errMissingMetadata = errors.New("metadata field missing or not a string")

// reportError forwards e to config.OnError, filling the subscription identifiers.
func (s *Subscriber) reportError(ctx context.Context, config eventbus.SubscriptionConfig, e eventbus.SubscriberError) {
	if config.OnError == nil {
		return
	}

	e.Stream = config.Stream
	e.ConsumerGroup = config.ConsumerGroup
	e.ConsumerID = config.ConsumerID

	// A faulty observer must not kill the processing goroutine or skip the ACK.
	defer func() { _ = recover() }()

	config.OnError(ctx, e)
}

// ack acknowledges msg and reports a failure to do so.
func (s *Subscriber) ack(
	ctx context.Context, config eventbus.SubscriptionConfig, msg redis.XMessage,
	eventID, eventType string, attempt int,
) {
	if err := s.client.XAck(ctx, config.Stream, config.ConsumerGroup, msg.ID).Err(); err != nil {
		s.reportError(ctx, config, eventbus.SubscriberError{
			Stage:     eventbus.StageAck,
			MessageID: msg.ID,
			EventID:   eventID,
			EventType: eventType,
			Attempt:   attempt,
			Err:       err,
		})
	}
}

// requeue re-adds msg to the stream with an incremented attempt counter.
func (s *Subscriber) requeue(
	ctx context.Context, config eventbus.SubscriptionConfig, msg redis.XMessage,
	metadata map[string]any, attempt int, eventID, eventType string,
) {
	metadata["attempt"] = attempt + 1

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		s.reportError(ctx, config, eventbus.SubscriberError{
			Stage:     eventbus.StageMetadataMarshal,
			MessageID: msg.ID,
			EventID:   eventID,
			EventType: eventType,
			Attempt:   attempt,
			Err:       err,
		})

		return
	}

	err = s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: config.Stream,
		Values: map[string]any{
			fieldMetadata: string(metadataJSON),
			fieldPayload:  msg.Values[fieldPayload],
		},
	}).Err()
	if err != nil {
		s.reportError(ctx, config, eventbus.SubscriberError{
			Stage:     eventbus.StageRetryRequeue,
			MessageID: msg.ID,
			EventID:   eventID,
			EventType: eventType,
			Attempt:   attempt,
			Err:       err,
		})
	}
}

// processMessage processes a single message with retry logic.
func (s *Subscriber) processMessage(ctx context.Context, config eventbus.SubscriptionConfig, msg redis.XMessage) {
	// Parse metadata
	var metadata map[string]any
	metadataStr, ok := msg.Values["metadata"].(string)
	if !ok {
		// Invalid message format, report and acknowledge to prevent reprocessing
		s.reportError(ctx, config, eventbus.SubscriberError{
			Stage: eventbus.StageMetadataUnmarshal, MessageID: msg.ID, Err: errMissingMetadata,
		})
		s.ack(ctx, config, msg, "", "", 0)

		return
	}

	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		// Invalid metadata, report and acknowledge to prevent reprocessing
		s.reportError(ctx, config, eventbus.SubscriberError{
			Stage: eventbus.StageMetadataUnmarshal, MessageID: msg.ID, Err: err,
		})
		s.ack(ctx, config, msg, "", "", 0)

		return
	}

	// Get attempt count
	attempt := 1
	if attemptVal, ok := metadata["attempt"].(float64); ok {
		attempt = int(attemptVal)
	}

	// Create timeout context for processing
	processCtx, cancel := context.WithTimeout(ctx, config.BlockDuration)
	defer cancel()

	// Call handler with raw event data
	// Note: In a real implementation, you would deserialize based on event type
	// For MVP, we pass the raw event through the handler
	// For now, we create a minimal event wrapper
	// Services will need to deserialize based on event type
	id, _ := metadata["id"].(string)
	eventType, _ := metadata["type"].(string)
	timestampStr, _ := metadata["timestamp"].(string)
	payload, _ := msg.Values["payload"].(string)

	event := &rawEvent{
		id:        id,
		eventType: eventType,
		timestamp: parseTime(timestampStr),
		data:      payload,
	}

	if handlerErr := config.Handler(processCtx, event); handlerErr != nil {
		// Handle retry logic
		if attempt < 3 { // Max 3 attempts
			time.Sleep(calculateBackoff(attempt))
			s.requeue(ctx, config, msg, metadata, attempt, id, eventType)
		} else if config.DLQPublisher != nil {
			dlqEntry := eventbus.NewDLQEntry(config.Stream, event, handlerErr, config.DLQService, attempt)
			if err := config.DLQPublisher.Publish(ctx, streams.StreamDLQ, dlqEntry); err != nil {
				s.reportError(ctx, config, eventbus.SubscriberError{
					Stage:     eventbus.StageDLQPublish,
					MessageID: msg.ID,
					EventID:   id,
					EventType: eventType,
					Attempt:   attempt,
					Err:       err,
				})
			}
		}
	}

	// Acknowledge in every outcome (success, retry requeued or lost, DLQ routed or dropped).
	s.ack(ctx, config, msg, id, eventType, attempt)
}

// calculateBackoff calculates exponential backoff.
func calculateBackoff(attempt int) time.Duration {
	if attempt <= 1 {
		return 0
	}

	// Compare in float64 milliseconds before casting to time.Duration: for large
	// attempt counts, 100*5^(attempt-2) overflows int64 once multiplied by
	// time.Millisecond, producing a negative garbage duration.
	const capMillis = 10_000.0

	backoffMillis := 100 * math.Pow(5, float64(attempt-2))
	if backoffMillis > capMillis || math.IsInf(backoffMillis, 1) {
		return 10 * time.Second
	}

	return time.Duration(backoffMillis) * time.Millisecond
}

// parseTime parses RFC3339 timestamp.
func parseTime(timestamp string) time.Time {
	t, _ := time.Parse(time.RFC3339, timestamp)

	return t
}

// Close closes the Redis connection.
func (s *Subscriber) Close() error {
	return s.client.Close()
}

// Health checks Redis connection health.
func (s *Subscriber) Health(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// rawEvent is a minimal event wrapper for raw event data.
type rawEvent struct {
	id        string
	eventType string
	timestamp time.Time
	data      string
}

func (e *rawEvent) EventType() string    { return e.eventType }
func (e *rawEvent) EventID() string      { return e.id }
func (e *rawEvent) EventTime() time.Time { return e.timestamp }
func (e *rawEvent) Data() string         { return e.data }
func (e *rawEvent) Validate() error      { return nil }
