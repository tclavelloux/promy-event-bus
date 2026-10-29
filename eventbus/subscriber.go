package eventbus

import (
	"context"
	"fmt"
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

	// DLQPublisher is an optional publisher used to route failed events to the
	// dead-letter queue after retry exhaustion. If nil, exhausted events are
	// dropped (still ACKed).
	DLQPublisher EventPublisher

	// DLQService identifies the service for DLQ entries (e.g., "promy-crm").
	DLQService string

	// OnError, if non-nil, is called synchronously from the processing goroutine whenever the subscriber
	// swallows an error (see SubscriberErrorStage). It reports; it never changes delivery: the message is
	// ACKed or left pending exactly as without it. Keep it cheap and non-blocking (log, increment a metric).
	// It may run concurrently when MaxConcurrency > 1. ctx is the Subscribe context and may already be done.
	// A panic inside OnError is recovered and discarded. Handler errors are not reported here: the handler already has them.
	OnError func(ctx context.Context, err SubscriberError)
}

// SubscriberErrorStage identifies the step of message processing where an otherwise swallowed error occurred.
type SubscriberErrorStage string

const (
	// StageMetadataUnmarshal: metadata field missing, not a string, or invalid JSON; message ACKed and dropped.
	StageMetadataUnmarshal SubscriberErrorStage = "metadata_unmarshal"
	// StageMetadataMarshal: re-encoding metadata for a retry failed; message ACKed and dropped.
	StageMetadataMarshal SubscriberErrorStage = "metadata_marshal"
	// StageRetryRequeue: XADD of the retry copy failed; original still ACKed, retry lost.
	StageRetryRequeue SubscriberErrorStage = "retry_requeue"
	// StageDLQPublish: DLQPublisher.Publish failed after retry exhaustion; message still ACKed, event lost.
	StageDLQPublish SubscriberErrorStage = "dlq_publish"
	// StageAck: XACK failed; message stays pending (no reclaim today).
	StageAck SubscriberErrorStage = "ack"
)

// SubscriberError describes an error the subscriber handled internally instead of returning.
type SubscriberError struct {
	// Stage is the processing step that failed.
	Stage SubscriberErrorStage
	// Stream is the consumed stream.
	Stream string
	// ConsumerGroup is the consumer group of the subscription.
	ConsumerGroup string
	// ConsumerID is the consumer within the group.
	ConsumerID string
	// MessageID is the Redis stream entry ID.
	MessageID string
	// EventID is the event ID; empty when metadata could not be decoded.
	EventID string
	// EventType is the event type; empty when metadata could not be decoded.
	EventType string
	// Attempt is the delivery attempt; 0 when metadata could not be decoded.
	Attempt int
	// Err is the underlying error.
	Err error
}

// Error implements the error interface.
func (e SubscriberError) Error() string {
	return fmt.Sprintf("eventbus subscriber: %s failed for message %s on %s: %v", e.Stage, e.MessageID, e.Stream, e.Err)
}

// Unwrap returns the underlying error.
func (e SubscriberError) Unwrap() error {
	return e.Err
}

// EventHandler processes a single event.
// Return nil if the event was processed successfully.
// Return an error to trigger retry logic.
type EventHandler func(ctx context.Context, event Event) error
