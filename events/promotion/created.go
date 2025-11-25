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

	// Required core identity fields
	PromotionID     string    `json:"promotion_id"`
	PromotionName   string    `json:"promotion_name"`
	DistributorID   string    `json:"distributor_id"`
	LeafletID       string    `json:"leaflet_id"`
	LeafletPage     int       `json:"leaflet_page"`
	DiscountedPrice float64   `json:"discounted_price"`
	CreatedAt       time.Time `json:"created_at"`

	// Optional fields - nullable to preserve source semantics
	ProductTypeID *string   `json:"product_type_id,omitempty"` // May not be identified yet
	Dates         *[]string `json:"dates,omitempty"`           // nil=pending, []=unrestricted, values=specific dates
	ImageURL      *string   `json:"image_url,omitempty"`       // May not have image
	OriginalPrice *float64  `json:"original_price,omitempty"`  // For discount calculation
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
	dates *[]string,
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
		Dates:           dates,
		ImageURL:        imageURL,
		OriginalPrice:   originalPrice,
		CreatedAt:       time.Now().UTC(),
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
