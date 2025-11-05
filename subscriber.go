package eventbus

import (
	"context"
	"time"
)

// EventSubscriber subscribes to events from streams.
// Implementations must be safe for concurrent use.
type EventSubscriber interface {
	// Subscribe starts consuming events from the specified configuration.
	// This is a blocking operation that runs until context is cancelled.
	Subscribe(ctx context.Context, config SubscriptionConfig) error

	// Close gracefully shuts down the subscriber and releases resources.
	Close() error

	// Health checks the connection health.
	Health(ctx context.Context) error
}

// SubscriptionConfig configures how events are consumed.
type SubscriptionConfig struct {
	// Stream is the name of the stream to consume from.
	Stream string

	// ConsumerGroup is the consumer group name.
	// Multiple consumers with the same group share the workload.
	ConsumerGroup string

	// ConsumerID uniquely identifies this consumer within the group.
	ConsumerID string

	// Handler is called for each event received.
	Handler EventHandler

	// MaxConcurrency limits concurrent event processing.
	// If 0, defaults to 1 (sequential processing).
	MaxConcurrency int

	// BatchSize is the number of events to fetch per read.
	// If 0, defaults to 1.
	BatchSize int

	// BlockDuration is how long to block waiting for events.
	// If 0, defaults to 1 second.
	BlockDuration time.Duration
}

// EventHandler processes a single event.
// Return nil if the event was processed successfully.
// Return an error to trigger retry logic.
type EventHandler func(ctx context.Context, event Event) error
