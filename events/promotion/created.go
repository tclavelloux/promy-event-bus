package promotion

import (
	"encoding/json"
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// PromotionCreatedEvent is published when a promotion is created.
type PromotionCreatedEvent struct {
	eventbus.BaseEvent

	// Required core identity fields
	PromotionID     string    `json:"promotion_id"`
	PromotionName   string    `json:"promotion_name"`
	DistributorID   string    `json:"distributor_id"`
	LeafletID       string    `json:"leaflet_id"`
	LeafletPage     int       `json:"leaflet_page"`
	DiscountedPrice float64   `json:"discounted_price"`
	CreatedAt       time.Time `json:"created_at"`

	// Optional fields - nullable to preserve source semantics
	ProductTypeID *string    `json:"product_type_id,omitempty"` // May not be identified yet
	ValidFrom     *time.Time `json:"valid_from,omitempty"`      // nil=pending, values=specific date
	ValidTo       *time.Time `json:"valid_to,omitempty"`        // nil=pending, values=specific date
	ImageURL      *string    `json:"image_url,omitempty"`       // May not have image
	OriginalPrice *float64   `json:"original_price,omitempty"`  // For discount calculation
}

// NewPromotionCreatedEvent creates a new promotion created event.
func NewPromotionCreatedEvent(
	promotionID string,
	promotionName string,
	distributorID string,
	leafletID string,
	leafletPage int,
	discountedPrice float64,
	productTypeID *string,
	validFrom *time.Time,
	validTo *time.Time,
	imageURL *string,
	originalPrice *float64,
) *PromotionCreatedEvent {
	return &PromotionCreatedEvent{
		BaseEvent:       eventbus.NewBaseEvent(events.EventPromotionCreated, "promy-product"),
		PromotionID:     promotionID,
		PromotionName:   promotionName,
		DistributorID:   distributorID,
		LeafletID:       leafletID,
		LeafletPage:     leafletPage,
		DiscountedPrice: discountedPrice,
		ProductTypeID:   productTypeID,
		ValidFrom:       validFrom,
		ValidTo:         validTo,
		ImageURL:        imageURL,
		OriginalPrice:   originalPrice,
		CreatedAt:       time.Now().UTC(),
	}
}

// Data returns the JSON payload of the event.
func (e *PromotionCreatedEvent) Data() string {
	b, _ := json.Marshal(e) //nolint:errchkjson // struct fields are always JSON-safe

	return string(b)
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
	if e.LeafletID == "" {
		return fmt.Errorf("%w: leaflet_id is required", eventbus.ErrInvalidEvent)
	}
	if e.LeafletPage <= 0 {
		return fmt.Errorf("%w: leaflet_page must be positive", eventbus.ErrInvalidEvent)
	}
	if e.DiscountedPrice <= 0 {
		return fmt.Errorf("%w: discounted_price must be positive", eventbus.ErrInvalidEvent)
	}

	return nil
}
