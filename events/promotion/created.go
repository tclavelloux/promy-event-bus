package promotion

import (
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// PromotionCreatedEvent is published when a promotion is created.
type PromotionCreatedEvent struct {
	eventbus.BaseEvent

	PromotionID   string    `json:"promotion_id"`
	PromotionName string    `json:"promotion_name"`
	DistributorID string    `json:"distributor_id"`
	ProductTypeID string    `json:"product_type_id"`
	Dates         []string  `json:"dates"`
	Price         float64   `json:"price"`
	ImageURL      string    `json:"image_url"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewPromotionCreatedEvent creates a new promotion created event.
func NewPromotionCreatedEvent(
	promotionID, promotionName, distributorID, productTypeID string,
	dates []string,
	price float64,
	imageURL string,
) *PromotionCreatedEvent {
	return &PromotionCreatedEvent{
		BaseEvent:     eventbus.NewBaseEvent(events.EventPromotionCreated, "promy-product"),
		PromotionID:   promotionID,
		PromotionName: promotionName,
		DistributorID: distributorID,
		ProductTypeID: productTypeID,
		Dates:         dates,
		Price:         price,
		ImageURL:      imageURL,
		CreatedAt:     time.Now().UTC(),
	}
}

// Validate validates the promotion created event.
func (e *PromotionCreatedEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}
	if e.PromotionID == "" {
		return fmt.Errorf("%w: promotion_id is required", eventbus.ErrInvalidEvent)
	}
	if e.PromotionName == "" {
		return fmt.Errorf("%w: promotion_name is required", eventbus.ErrInvalidEvent)
	}
	if e.DistributorID == "" {
		return fmt.Errorf("%w: distributor_id is required", eventbus.ErrInvalidEvent)
	}

	return nil
}
