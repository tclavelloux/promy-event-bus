package user

import (
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// UserLocationUpdatedEvent is published when user location is updated.
type UserLocationUpdatedEvent struct {
	eventbus.BaseEvent

	UserID    string    `json:"user_id" validate:"required,min=1"`
	Latitude  float64   `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64   `json:"longitude" validate:"required,min=-180,max=180"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

// NewUserLocationUpdatedEvent creates a new user location updated event.
func NewUserLocationUpdatedEvent(userID string, latitude, longitude float64) *UserLocationUpdatedEvent {
	return &UserLocationUpdatedEvent{
		BaseEvent: eventbus.NewBaseEvent(events.EventUserLocationUpdated, "promy-user"),
		UserID:    userID,
		Latitude:  latitude,
		Longitude: longitude,
		UpdatedAt: time.Now().UTC(),
	}
}

// Validate validates the user location updated event using struct tags.
func (e *UserLocationUpdatedEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}

	// Validation happens automatically via struct tags
	// The validator is injected by the publisher/subscriber
	return nil
}
