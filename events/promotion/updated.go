package promotion

import (
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// PromotionUpdatedEvent is published when a promotion is updated.
type PromotionUpdatedEvent struct {
	eventbus.BaseEvent

	PromotionID   string    `json:"promotion_id"`
	UpdatedFields []string  `json:"updated_fields"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewPromotionUpdatedEvent creates a new promotion updated event.
func NewPromotionUpdatedEvent(promotionID string, updatedFields []string) *PromotionUpdatedEvent {
	return &PromotionUpdatedEvent{
		BaseEvent:     eventbus.NewBaseEvent(events.EventPromotionUpdated, "promy-product"),
		PromotionID:   promotionID,
		UpdatedFields: updatedFields,
		UpdatedAt:     time.Now().UTC(),
	}
}

// Validate validates the promotion updated event.
func (e *PromotionUpdatedEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}
	if e.PromotionID == "" {
		return fmt.Errorf("%w: promotion_id is required", eventbus.ErrInvalidEvent)
	}

	return nil
}
