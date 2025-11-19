package promotion

import (
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// PromotionDeletedEvent is published when a promotion is deleted.
type PromotionDeletedEvent struct {
	eventbus.BaseEvent

	PromotionID string    `json:"promotion_id"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// NewPromotionDeletedEvent creates a new promotion deleted event.
func NewPromotionDeletedEvent(promotionID string) *PromotionDeletedEvent {
	return &PromotionDeletedEvent{
		BaseEvent:   eventbus.NewBaseEvent(events.EventPromotionDeleted, "promy-product"),
		PromotionID: promotionID,
		DeletedAt:   time.Now().UTC(),
	}
}

// Validate validates the promotion deleted event.
func (e *PromotionDeletedEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}
	if e.PromotionID == "" {
		return fmt.Errorf("%w: promotion_id is required", eventbus.ErrInvalidEvent)
	}

	return nil
}
