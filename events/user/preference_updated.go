package user

import (
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

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
		BaseEvent:            eventbus.NewBaseEvent(events.EventUserPreferencesUpdated, "promy-user"),
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
