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

// processMessage processes a single message with retry logic.
func (s *Subscriber) processMessage(ctx context.Context, config eventbus.SubscriptionConfig, msg redis.XMessage) {
	// Parse metadata
	var metadata map[string]any
	metadataStr, ok := msg.Values["metadata"].(string)
	if !ok {
		// Invalid message format, acknowledge to prevent reprocessing
		s.client.XAck(ctx, config.Stream, config.ConsumerGroup, msg.ID)

		return
	}

	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		// Invalid metadata, acknowledge to prevent reprocessing
		s.client.XAck(ctx, config.Stream, config.ConsumerGroup, msg.ID)

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

	if err := config.Handler(processCtx, event); err != nil {
		// Handle retry logic
		if attempt < 3 { // Max 3 attempts
			// Calculate backoff
			backoff := calculateBackoff(attempt)
			time.Sleep(backoff)

			// Increment attempt and re-publish for retry
			metadata["attempt"] = attempt + 1
			metadataJSON, err := json.Marshal(metadata)
			if err != nil {
				// Log error but continue (acknowledge message)
				return
			}

			// Re-add to stream for retry
			s.client.XAdd(ctx, &redis.XAddArgs{
				Stream: config.Stream,
				Values: map[string]any{
					"metadata": string(metadataJSON),
					"payload":  msg.Values["payload"],
				},
			})
		}
		// After max retries, message is lost (DLQ would go here in future)
		// For now, just acknowledge to prevent infinite loop
		s.client.XAck(ctx, config.Stream, config.ConsumerGroup, msg.ID)

		return
	}

	// Acknowledge successful processing
	s.client.XAck(ctx, config.Stream, config.ConsumerGroup, msg.ID)
}

// calculateBackoff calculates exponential backoff.
func calculateBackoff(attempt int) time.Duration {
	if attempt <= 1 {
		return 0
	}

	backoff := time.Duration(100*math.Pow(5, float64(attempt-2))) * time.Millisecond

	// Cap at 10 seconds
	if backoff > 10*time.Second {
		return 10 * time.Second
	}

	return backoff
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
func (e *rawEvent) Validate() error      { return nil }
