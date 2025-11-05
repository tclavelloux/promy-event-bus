package eventbus

import "errors"

// Domain-specific errors for event bus operations.
var (
	// ErrInvalidEvent is returned when an event fails validation.
	ErrInvalidEvent = errors.New("invalid event")

	// ErrPublishFailed is returned when event publishing fails.
	ErrPublishFailed = errors.New("failed to publish event")

	// ErrSubscriptionFailed is returned when subscription fails.
	ErrSubscriptionFailed = errors.New("failed to subscribe")

	// ErrConnectionClosed is returned when the connection is closed.
	ErrConnectionClosed = errors.New("connection closed")

	// ErrConsumerGroupExists is returned when attempting to create an existing consumer group.
	ErrConsumerGroupExists = errors.New("consumer group already exists")
)
