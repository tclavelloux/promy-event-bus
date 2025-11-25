package product

import (
	"fmt"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// ProductIdentifiedEvent is published when a product is identified by the AI system.
type ProductIdentifiedEvent struct {
	eventbus.BaseEvent

	PromotionID string  `json:"promotion_id"`
	ProductID   string  `json:"product_id"`
	ProductType string  `json:"product_type"`
	CategoryID  string  `json:"category_id"`
	Brand       *string `json:"brand,omitempty"`
	Confidence  float64 `json:"confidence"`
}

// NewProductIdentifiedEvent creates a new product identified event.
func NewProductIdentifiedEvent(
	promotionID string,
	productID string,
	productType string,
	categoryID string,
	brand *string,
	confidence float64,
) *ProductIdentifiedEvent {
	return &ProductIdentifiedEvent{
		BaseEvent:   eventbus.NewBaseEvent(events.EventProductIdentified, "promy-identifier"),
		PromotionID: promotionID,
		ProductID:   productID,
		ProductType: productType,
		CategoryID:  categoryID,
		Brand:       brand,
		Confidence:  confidence,
	}
}

// Validate validates the product identified event.
func (e *ProductIdentifiedEvent) Validate() error {
	if err := e.BaseEvent.Validate(); err != nil {
		return err
	}
	if e.PromotionID == "" {
		return fmt.Errorf("%w: promotion_id is required", eventbus.ErrInvalidEvent)
	}
	if e.ProductID == "" {
		return fmt.Errorf("%w: product_id is required", eventbus.ErrInvalidEvent)
	}
	if e.ProductType == "" {
		return fmt.Errorf("%w: product_type is required", eventbus.ErrInvalidEvent)
	}
	if e.CategoryID == "" {
		return fmt.Errorf("%w: category_id is required", eventbus.ErrInvalidEvent)
	}
	if e.Confidence < 0 || e.Confidence > 1 {
		return fmt.Errorf("%w: confidence must be between 0 and 1", eventbus.ErrInvalidEvent)
	}

	return nil
}
