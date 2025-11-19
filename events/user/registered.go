package user

import (
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// UserRegisteredEvent is published when a new user registers.
type UserRegisteredEvent struct {
	eventbus.BaseEvent

	UserID       string    `json:"user_id" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	RegisteredAt time.Time `json:"registered_at" validate:"required"`
}

// NewUserRegisteredEvent creates a new user registered event.
func NewUserRegisteredEvent(userID, email string) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		BaseEvent:    eventbus.NewBaseEvent(events.EventUserRegistered, "promy-user"),
		UserID:       userID,
		Email:        email,
		RegisteredAt: time.Now().UTC(),
	}
}

// Validate validates the user registered event using struct tags.
func (e *UserRegisteredEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}

	// Validation happens automatically via struct tags
	// The validator is injected by the publisher/subscriber
	return nil
}
