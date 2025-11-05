package events

import (
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus"
)

// UserRegisteredEvent is published when a new user registers.
type UserRegisteredEvent struct {
	eventbus.BaseEvent

	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	RegisteredAt time.Time `json:"registered_at"`
}

// NewUserRegisteredEvent creates a new user registered event.
func NewUserRegisteredEvent(userID, email string) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		BaseEvent:    eventbus.NewBaseEvent(EventUserRegistered, "promy-user"),
		UserID:       userID,
		Email:        email,
		RegisteredAt: time.Now().UTC(),
	}
}

// Validate validates the user registered event.
func (e *UserRegisteredEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}
	if e.UserID == "" {
		return fmt.Errorf("%w: user_id is required", eventbus.ErrInvalidEvent)
	}
	if e.Email == "" {
		return fmt.Errorf("%w: email is required", eventbus.ErrInvalidEvent)
	}
	return nil
}

// UserPreferencesUpdatedEvent is published when user preferences are updated.
type UserPreferencesUpdatedEvent struct {
	eventbus.BaseEvent

	UserID               string    `json:"user_id"`
	FavoriteDistributors []string  `json:"favorite_distributors"`
	FavoriteCategories   []string  `json:"favorite_categories"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// NewUserPreferencesUpdatedEvent creates a new user preferences updated event.
func NewUserPreferencesUpdatedEvent(
	userID string,
	favoriteDistributors, favoriteCategories []string,
) *UserPreferencesUpdatedEvent {
	return &UserPreferencesUpdatedEvent{
		BaseEvent:            eventbus.NewBaseEvent(EventUserPreferencesUpdated, "promy-user"),
		UserID:               userID,
		FavoriteDistributors: favoriteDistributors,
		FavoriteCategories:   favoriteCategories,
		UpdatedAt:            time.Now().UTC(),
	}
}

// Validate validates the user preferences updated event.
func (e *UserPreferencesUpdatedEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}
	if e.UserID == "" {
		return fmt.Errorf("%w: user_id is required", eventbus.ErrInvalidEvent)
	}
	return nil
}

// UserLocationUpdatedEvent is published when user location is updated.
type UserLocationUpdatedEvent struct {
	eventbus.BaseEvent

	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUserLocationUpdatedEvent creates a new user location updated event.
func NewUserLocationUpdatedEvent(userID string, latitude, longitude float64) *UserLocationUpdatedEvent {
	return &UserLocationUpdatedEvent{
		BaseEvent: eventbus.NewBaseEvent(EventUserLocationUpdated, "promy-user"),
		UserID:    userID,
		Latitude:  latitude,
		Longitude: longitude,
		UpdatedAt: time.Now().UTC(),
	}
}

// Validate validates the user location updated event.
func (e *UserLocationUpdatedEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}
	if e.UserID == "" {
		return fmt.Errorf("%w: user_id is required", eventbus.ErrInvalidEvent)
	}
	if e.Latitude < -90 || e.Latitude > 90 {
		return fmt.Errorf("%w: latitude must be between -90 and 90", eventbus.ErrInvalidEvent)
	}
	if e.Longitude < -180 || e.Longitude > 180 {
		return fmt.Errorf("%w: longitude must be between -180 and 180", eventbus.ErrInvalidEvent)
	}
	return nil
}
