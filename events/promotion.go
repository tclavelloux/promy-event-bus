package events

import (
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus"
)

// PromotionCreatedEvent is published when a promotion is created.
type PromotionCreatedEvent struct {
	eventbus.BaseEvent

	PromotionID   string    `json:"promotion_id"`
	PromotionName string    `json:"promotion_name"`
	DistributorID string    `json:"distributor_id"`
	CategoryID    string    `json:"category_id"`
	Dates         []string  `json:"dates"`
	Price         float64   `json:"price"`
	ImageURL      string    `json:"image_url"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewPromotionCreatedEvent creates a new promotion created event.
func NewPromotionCreatedEvent(
	promotionID, promotionName, distributorID, categoryID string,
	dates []string,
	price float64,
	imageURL string,
) *PromotionCreatedEvent {
	return &PromotionCreatedEvent{
		BaseEvent:     eventbus.NewBaseEvent(EventPromotionCreated, "promy-product"),
		PromotionID:   promotionID,
		PromotionName: promotionName,
		DistributorID: distributorID,
		CategoryID:    categoryID,
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
		BaseEvent:     eventbus.NewBaseEvent(EventPromotionUpdated, "promy-product"),
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

// PromotionDeletedEvent is published when a promotion is deleted.
type PromotionDeletedEvent struct {
	eventbus.BaseEvent

	PromotionID string    `json:"promotion_id"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// NewPromotionDeletedEvent creates a new promotion deleted event.
func NewPromotionDeletedEvent(promotionID string) *PromotionDeletedEvent {
	return &PromotionDeletedEvent{
		BaseEvent:   eventbus.NewBaseEvent(EventPromotionDeleted, "promy-product"),
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
