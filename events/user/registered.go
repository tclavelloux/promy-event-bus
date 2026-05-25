package user

import (
	"encoding/json"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// UserRegisteredEvent is published when a new user registers.
type UserRegisteredEvent struct {
	eventbus.BaseEvent

	UserID string `json:"user_id" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
}

// NewUserRegisteredEvent creates a new user registered event.
func NewUserRegisteredEvent(userID, email string) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		BaseEvent: eventbus.NewBaseEvent(events.EventUserRegistered, "promy-user"),
		UserID:    userID,
		Email:     email,
	}
}

// Data returns the JSON payload of the event.
func (e *UserRegisteredEvent) Data() string {
	b, _ := json.Marshal(e) //nolint:errchkjson // struct fields are always JSON-safe

	return string(b)
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
