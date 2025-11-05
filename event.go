package eventbus

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event that can be published and consumed.
type Event interface {
	// EventType returns the event type (e.g., "promotion.created").
	EventType() string

	// EventID returns a unique identifier for this event.
	EventID() string

	// EventTime returns when the event occurred.
	EventTime() time.Time

	// Validate checks if the event is valid.
	Validate() error
}

// BaseEvent provides common event fields that all events should embed.
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Source    string    `json:"source"`
}

// NewBaseEvent creates a new base event with generated ID and current timestamp.
func NewBaseEvent(eventType, source string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Version:   "1.0",
		Source:    source,
	}
}

// EventType returns the event type.
func (e BaseEvent) EventType() string {
	return e.Type
}

// EventID returns the event ID.
func (e BaseEvent) EventID() string {
	return e.ID
}

// EventTime returns the event timestamp.
func (e BaseEvent) EventTime() time.Time {
	return e.Timestamp
}

// Validate performs basic validation on the base event.
func (e BaseEvent) Validate() error {
	if e.ID == "" {
		return ErrInvalidEvent
	}
	if e.Type == "" {
		return ErrInvalidEvent
	}
	if e.Timestamp.IsZero() {
		return ErrInvalidEvent
	}
	return nil
}
