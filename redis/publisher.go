package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus"

	"github.com/redis/go-redis/v9"
)

// Publisher implements EventPublisher for Redis Streams.
type Publisher struct {
	client *redis.Client
	config eventbus.RedisConfig
}

// NewPublisher creates a new Redis publisher.
func NewPublisher(config eventbus.RedisConfig) (*Publisher, error) {
	opts, err := redis.ParseURL(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis DSN: %w", err)
	}

	// Apply configuration
	if config.PoolSize > 0 {
		opts.PoolSize = config.PoolSize
	}
	if config.MaxRetries > 0 {
		opts.MaxRetries = config.MaxRetries
	}
	if config.MinRetryBackoff > 0 {
		opts.MinRetryBackoff = config.MinRetryBackoff
	}
	if config.MaxRetryBackoff > 0 {
		opts.MaxRetryBackoff = config.MaxRetryBackoff
	}
	if config.DialTimeout > 0 {
		opts.DialTimeout = config.DialTimeout
	}
	if config.ReadTimeout > 0 {
		opts.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout > 0 {
		opts.WriteTimeout = config.WriteTimeout
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Publisher{
		client: client,
		config: config,
	}, nil
}

// Publish publishes a single event to Redis Streams.
func (p *Publisher) Publish(ctx context.Context, stream string, event eventbus.Event) error {
	// Validate event
	if err := event.Validate(); err != nil {
		return fmt.Errorf("%w: %w", eventbus.ErrInvalidEvent, err)
	}

	// Serialize metadata
	metadata := map[string]any{
		"id":        event.EventID(),
		"type":      event.EventType(),
		"timestamp": event.EventTime().Format(time.RFC3339),
		"version":   "1.0",
		"attempt":   1,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Serialize payload
	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Publish to Redis Stream
	args := &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"metadata": string(metadataJSON),
			"payload":  string(payloadJSON),
		},
	}

	if err := p.client.XAdd(ctx, args).Err(); err != nil {
		return fmt.Errorf("%w: %w", eventbus.ErrPublishFailed, err)
	}

	return nil
}

// PublishBatch publishes multiple events in a pipeline.
func (p *Publisher) PublishBatch(ctx context.Context, stream string, events []eventbus.Event) error {
	if len(events) == 0 {
		return nil
	}

	pipe := p.client.Pipeline()

	for _, event := range events {
		if err := event.Validate(); err != nil {
			return fmt.Errorf("%w: event %s: %w", eventbus.ErrInvalidEvent, event.EventID(), err)
		}

		metadata := map[string]any{
			"id":        event.EventID(),
			"type":      event.EventType(),
			"timestamp": event.EventTime().Format(time.RFC3339),
			"version":   "1.0",
			"attempt":   1,
		}
		metadataJSON, _ := json.Marshal(metadata)
		payloadJSON, _ := json.Marshal(event)

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]any{
				"metadata": string(metadataJSON),
				"payload":  string(payloadJSON),
			},
		})
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", eventbus.ErrPublishFailed, err)
	}

	return nil
}

// Close closes the Redis connection.
func (p *Publisher) Close() error {
	return p.client.Close()
}

// Health checks Redis connection health.
func (p *Publisher) Health(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}
