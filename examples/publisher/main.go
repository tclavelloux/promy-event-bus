//nolint:all // Example file
package main

import (
	"context"
	"log"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
	"github.com/tclavelloux/promy-event-bus/streams"
	"github.com/tclavelloux/promy-event-bus/testutil"
)

func main() {
	log.Println("Starting Event Bus Publisher Example...")

	config := eventbus.RedisConfig{
		DSN:      "redis://localhost:6379/0",
		PoolSize: 10,
	}

	publisher, err := redis.NewPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Failed to close publisher: %v", err)
		}
	}()

	log.Println("Publisher created successfully")

	ctx := context.Background()

	// Example 1: Publish a single promotion event
	log.Println("\n=== Publishing Single Promotion Event ===")
	promoEvent := testutil.NewTestEvent("promotion.created", map[string]any{
		"promotion_id":   "promo-123",
		"promotion_name": "Patate douce",
		"distributor_id": "dist-456",
	})

	if err := publisher.Publish(ctx, streams.StreamPromotions, promoEvent); err != nil {
		log.Printf("Failed to publish promotion event: %v", err)
	} else {
		log.Printf("Published promotion event (ID: %s)", promoEvent.EventID())
	}

	// Example 2: Publish a user registered event
	log.Println("\n=== Publishing User Registered Event ===")
	userEvent := testutil.NewTestEvent("user.registered", map[string]any{
		"user_id": "user-789",
		"email":   "john.doe@example.com",
	})

	if err := publisher.Publish(ctx, streams.StreamUsers, userEvent); err != nil {
		log.Printf("Failed to publish user event: %v", err)
	} else {
		log.Printf("Published user event (ID: %s)", userEvent.EventID())
	}

	// Example 3: Publish a product identified event
	log.Println("\n=== Publishing Product Identified Event ===")
	productEvent := testutil.NewTestEvent("product.identified", map[string]any{
		"promotion_id": "promo-123",
		"product_id":   "prod-456",
		"product_type": "vegetables",
		"category_id":  "cat-789",
		"confidence":   0.95,
	})

	if err := publisher.Publish(ctx, streams.StreamProducts, productEvent); err != nil {
		log.Printf("Failed to publish product event: %v", err)
	} else {
		log.Printf("Published product event (ID: %s)", productEvent.EventID())
	}

	// Example 4: Batch publish
	log.Println("\n=== Publishing Batch of Events ===")
	batchEvents := []eventbus.Event{
		testutil.NewTestEvent("promotion.created", map[string]any{"promotion_id": "promo-200", "promotion_name": "Pommes Golden"}),
		testutil.NewTestEvent("promotion.created", map[string]any{"promotion_id": "promo-201", "promotion_name": "Bananes"}),
		testutil.NewTestEvent("promotion.created", map[string]any{"promotion_id": "promo-202", "promotion_name": "Oranges"}),
	}

	if err := publisher.PublishBatch(ctx, streams.StreamPromotions, batchEvents); err != nil {
		log.Printf("Failed to publish batch: %v", err)
	} else {
		log.Printf("Published batch of %d promotion events", len(batchEvents))
	}

	// Example 5: Health check
	log.Println("\n=== Health Check ===")
	if err := publisher.Health(ctx); err != nil {
		log.Printf("Publisher health check failed: %v", err)
	} else {
		log.Println("Publisher is healthy")
	}

	time.Sleep(500 * time.Millisecond)
	log.Println("\n=== Publisher Example Complete ===")
}
