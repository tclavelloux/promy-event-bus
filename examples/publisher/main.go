package main

import (
	"context"
	"log"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus"
	"github.com/tclavelloux/promy-event-bus/events"
	"github.com/tclavelloux/promy-event-bus/redis"
)

func main() {
	log.Println("Starting Event Bus Publisher Example...")

	// Load configuration
	config := eventbus.RedisConfig{
		DSN:      "redis://localhost:6379/0",
		PoolSize: 10,
	}

	// Create publisher
	publisher, err := redis.NewPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	log.Println("Publisher created successfully")

	// Create context
	ctx := context.Background()

	// Example 1: Publish a single promotion event
	log.Println("\n=== Publishing Single Promotion Event ===")
	promoEvent := events.NewPromotionCreatedEvent(
		"promo-123",
		"Patate douce",
		"dist-456",
		"cat-789",
		[]string{"2025-11-06", "2025-11-07"},
		9.99,
		"https://example.com/patate-douce.jpg",
	)

	if err := publisher.Publish(ctx, events.StreamPromotions, promoEvent); err != nil {
		log.Printf("Failed to publish promotion event: %v", err)
	} else {
		log.Printf("✓ Published promotion event: %s (ID: %s)", promoEvent.PromotionName, promoEvent.EventID())
	}

	// Example 2: Publish a user registered event
	log.Println("\n=== Publishing User Registered Event ===")
	userEvent := events.NewUserRegisteredEvent("user-789", "john.doe@example.com")

	if err := publisher.Publish(ctx, events.StreamUsers, userEvent); err != nil {
		log.Printf("Failed to publish user event: %v", err)
	} else {
		log.Printf("✓ Published user event: %s (ID: %s)", userEvent.Email, userEvent.EventID())
	}

	// Example 3: Publish a product identified event
	log.Println("\n=== Publishing Product Identified Event ===")
	productEvent := events.NewProductIdentifiedEvent(
		"promo-123",
		"prod-456",
		"vegetables",
		"cat-789",
		"BioMarket",
		0.95,
	)

	if err := publisher.Publish(ctx, events.StreamProducts, productEvent); err != nil {
		log.Printf("Failed to publish product event: %v", err)
	} else {
		log.Printf("✓ Published product event: %s (ID: %s, Confidence: %.2f)", productEvent.ProductType, productEvent.EventID(), productEvent.Confidence)
	}

	// Example 4: Batch publish multiple events
	log.Println("\n=== Publishing Batch of Events ===")
	batchEvents := []eventbus.Event{
		events.NewPromotionCreatedEvent(
			"promo-200",
			"Pommes Golden",
			"dist-101",
			"cat-fruits",
			[]string{"2025-11-06"},
			3.99,
			"https://example.com/pommes.jpg",
		),
		events.NewPromotionCreatedEvent(
			"promo-201",
			"Bananes",
			"dist-101",
			"cat-fruits",
			[]string{"2025-11-06"},
			2.49,
			"https://example.com/bananes.jpg",
		),
		events.NewPromotionCreatedEvent(
			"promo-202",
			"Oranges",
			"dist-101",
			"cat-fruits",
			[]string{"2025-11-06"},
			4.99,
			"https://example.com/oranges.jpg",
		),
	}

	if err := publisher.PublishBatch(ctx, events.StreamPromotions, batchEvents); err != nil {
		log.Printf("Failed to publish batch: %v", err)
	} else {
		log.Printf("✓ Published batch of %d promotion events", len(batchEvents))
	}

	// Example 5: Health check
	log.Println("\n=== Health Check ===")
	if err := publisher.Health(ctx); err != nil {
		log.Printf("Publisher health check failed: %v", err)
	} else {
		log.Println("✓ Publisher is healthy")
	}

	// Wait a moment for events to be published
	time.Sleep(500 * time.Millisecond)

	log.Println("\n=== Publisher Example Complete ===")
	log.Println("Events have been published to Redis Streams")
	log.Println("Run the subscriber example to consume these events")
}
