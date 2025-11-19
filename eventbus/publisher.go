package eventbus

import "context"

// EventPublisher publishes events to a stream.
// Implementations must be safe for concurrent use.
type EventPublisher interface {
	// Publish publishes a single event to the specified stream.
	// This is a fire-and-forget operation that should not block.
	Publish(ctx context.Context, stream string, event Event) error

	// PublishBatch publishes multiple events atomically to the specified stream.
	// All events are published or none are published.
	PublishBatch(ctx context.Context, stream string, events []Event) error

	// Close gracefully shuts down the publisher and releases resources.
	Close() error

	// Health checks the connection health.
	Health(ctx context.Context) error
}
