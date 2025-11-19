//nolint:all // Example file
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
	"github.com/tclavelloux/promy-event-bus/redis"
)

func main() {
	log.Println("Starting Event Bus Subscriber Example...")

	// Load configuration
	config := eventbus.Config{
		Redis: eventbus.RedisConfig{
			DSN:      "redis://localhost:6379/0",
			PoolSize: 10,
		},
		Consumer: eventbus.ConsumerConfig{
			Group:          "example-consumers",
			ConsumerID:     "worker-1",
			BatchSize:      10,
			BlockDuration:  2 * time.Second,
			MaxConcurrency: 5,
		},
	}

	// Create subscriber
	subscriber, err := redis.NewSubscriber(config)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}
	defer subscriber.Close() //nolint:errcheck // Example cleanup

	log.Println("Subscriber created successfully")

	// Define event handler
	handler := func(ctx context.Context, event eventbus.Event) error {
		log.Printf("\n=== Received Event ===")
		log.Printf("Type: %s", event.EventType())
		log.Printf("ID: %s", event.EventID())
		log.Printf("Time: %s", event.EventTime().Format(time.RFC3339))

		// Type-specific handling (note: in real implementation you'd deserialize based on type)
		switch event.EventType() {
		case events.EventPromotionCreated:
			log.Printf("Event: New promotion created")
			// In real implementation: cast to *events.PromotionCreatedEvent and access fields

		case events.EventPromotionUpdated:
			log.Printf("Event: Promotion updated")

		case events.EventPromotionDeleted:
			log.Printf("Event: Promotion deleted")

		case events.EventProductIdentified:
			log.Printf("Event: Product identified by AI")

		case events.EventUserRegistered:
			log.Printf("Event: New user registered")

		case events.EventUserPreferencesUpdated:
			log.Printf("Event: User preferences updated")

		case events.EventUserLocationUpdated:
			log.Printf("Event: User location updated")

		default:
			log.Printf("Event: Unknown type")
		}

		// Simulate processing
		time.Sleep(100 * time.Millisecond)

		log.Printf("✓ Event processed successfully\n")
		return nil
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\nReceived signal %v, shutting down gracefully...", sig)
		cancel()
	}()

	// Start consuming from multiple streams
	log.Println("\n=== Starting Event Consumption ===")
	log.Println("Listening for events on:")
	log.Printf("  - %s (Promotion events)", events.StreamPromotions)
	log.Printf("  - %s (Product events)", events.StreamProducts)
	log.Printf("  - %s (User events)", events.StreamUsers)
	log.Println("\nPress Ctrl+C to stop...")

	// In a real application, you'd typically subscribe to one stream per subscriber
	// For this example, we'll subscribe to the promotions stream
	// You would spawn multiple goroutines or separate processes for other streams

	streamConfig := eventbus.SubscriptionConfig{
		Stream:         events.StreamPromotions,
		ConsumerGroup:  config.Consumer.Group,
		ConsumerID:     config.Consumer.ConsumerID,
		Handler:        handler,
		BatchSize:      config.Consumer.BatchSize,
		BlockDuration:  config.Consumer.BlockDuration,
		MaxConcurrency: config.Consumer.MaxConcurrency,
	}

	// Subscribe (blocking operation)
	if err := subscriber.Subscribe(ctx, streamConfig); err != nil {
		if err == context.Canceled {
			log.Println("\n=== Subscriber Stopped ===")
			log.Println("Graceful shutdown complete")
		} else {
			log.Fatalf("Subscription failed: %v", err)
		}
	}
}

// Example of how to subscribe to multiple streams in a real application
//
//nolint:all // Example function
func exampleMultiStreamSubscription() {
	config := eventbus.Config{
		Redis: eventbus.RedisConfig{
			DSN:      "redis://localhost:6379/0",
			PoolSize: 20,
		},
		Consumer: eventbus.ConsumerConfig{
			Group:          "multi-stream-consumers",
			ConsumerID:     fmt.Sprintf("worker-%d", time.Now().Unix()),
			BatchSize:      50,
			BlockDuration:  2 * time.Second,
			MaxConcurrency: 10,
		},
	}

	// Create separate subscribers for each stream
	promotionSubscriber, _ := redis.NewSubscriber(config)
	defer promotionSubscriber.Close()

	productSubscriber, _ := redis.NewSubscriber(config)
	defer productSubscriber.Close()

	userSubscriber, _ := redis.NewSubscriber(config)
	defer userSubscriber.Close()

	ctx := context.Background()

	// Subscribe to promotions stream
	go func() {
		promotionSubscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:         events.StreamPromotions,
			ConsumerGroup:  config.Consumer.Group,
			ConsumerID:     config.Consumer.ConsumerID,
			Handler:        handlePromotionEvent,
			BatchSize:      config.Consumer.BatchSize,
			BlockDuration:  config.Consumer.BlockDuration,
			MaxConcurrency: config.Consumer.MaxConcurrency,
		})
	}()

	// Subscribe to products stream
	go func() {
		productSubscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:         events.StreamProducts,
			ConsumerGroup:  config.Consumer.Group,
			ConsumerID:     config.Consumer.ConsumerID + "-products",
			Handler:        handleProductEvent,
			BatchSize:      config.Consumer.BatchSize,
			BlockDuration:  config.Consumer.BlockDuration,
			MaxConcurrency: config.Consumer.MaxConcurrency,
		})
	}()

	// Subscribe to users stream
	go func() {
		userSubscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:         events.StreamUsers,
			ConsumerGroup:  config.Consumer.Group,
			ConsumerID:     config.Consumer.ConsumerID + "-users",
			Handler:        handleUserEvent,
			BatchSize:      config.Consumer.BatchSize,
			BlockDuration:  config.Consumer.BlockDuration,
			MaxConcurrency: config.Consumer.MaxConcurrency,
		})
	}()
}

//nolint:unused // Example function
func handlePromotionEvent(ctx context.Context, event eventbus.Event) error {
	log.Printf("Processing promotion event: %s", event.EventID())
	return nil
}

//nolint:unused // Example function
func handleProductEvent(ctx context.Context, event eventbus.Event) error {
	log.Printf("Processing product event: %s", event.EventID())
	return nil
}

//nolint:unused // Example function
func handleUserEvent(ctx context.Context, event eventbus.Event) error {
	log.Printf("Processing user event: %s", event.EventID())
	return nil
}
