//nolint:all // Example file
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
	"github.com/tclavelloux/promy-event-bus/streams"
)

func main() {
	log.Println("Starting Event Bus Subscriber Example...")

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

	subscriber, err := redis.NewSubscriber(config)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}
	defer subscriber.Close() //nolint:errcheck // Example cleanup

	log.Println("Subscriber created successfully")

	handler := func(ctx context.Context, event eventbus.Event) error {
		log.Printf("\n=== Received Event ===")
		log.Printf("Type: %s", event.EventType())
		log.Printf("ID: %s", event.EventID())
		log.Printf("Time: %s", event.EventTime().Format(time.RFC3339))

		// Deserialize payload using event.Data()
		var payload map[string]any
		if err := json.Unmarshal([]byte(event.Data()), &payload); err != nil {
			log.Printf("Failed to unmarshal payload: %v", err)
		} else {
			log.Printf("Payload: %v", payload)
		}

		switch event.EventType() {
		case "promotion.created":
			log.Printf("Event: New promotion created")
		case "product.identified":
			log.Printf("Event: Product identified by AI")
		case "user.registered":
			log.Printf("Event: New user registered")
		case "user.preferences.updated":
			log.Printf("Event: User preferences updated")
		case "user.location.updated":
			log.Printf("Event: User location updated")
		default:
			log.Printf("Event: Unknown type -- ACKing")
		}

		log.Printf("Event processed successfully\n")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\nReceived signal %v, shutting down gracefully...", sig)
		cancel()
	}()

	log.Println("\n=== Starting Event Consumption ===")
	log.Printf("Listening on: %s", streams.StreamPromotions)
	log.Println("Press Ctrl+C to stop...")

	streamConfig := eventbus.SubscriptionConfig{
		Stream:         streams.StreamPromotions,
		ConsumerGroup:  config.Consumer.Group,
		ConsumerID:     config.Consumer.ConsumerID,
		Handler:        handler,
		BatchSize:      config.Consumer.BatchSize,
		BlockDuration:  config.Consumer.BlockDuration,
		MaxConcurrency: config.Consumer.MaxConcurrency,
	}

	if err := subscriber.Subscribe(ctx, streamConfig); err != nil {
		if err == context.Canceled {
			log.Println("\n=== Subscriber Stopped ===")
		} else {
			log.Fatalf("Subscription failed: %v", err)
		}
	}
}

// exampleMultiStreamSubscription shows how to subscribe to multiple streams.
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

	promotionSubscriber, _ := redis.NewSubscriber(config)
	defer promotionSubscriber.Close()

	userSubscriber, _ := redis.NewSubscriber(config)
	defer userSubscriber.Close()

	ctx := context.Background()

	handler := func(ctx context.Context, event eventbus.Event) error {
		log.Printf("Processing event: %s (type: %s)", event.EventID(), event.EventType())
		return nil
	}

	go func() {
		promotionSubscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:         streams.StreamPromotions,
			ConsumerGroup:  config.Consumer.Group,
			ConsumerID:     config.Consumer.ConsumerID,
			Handler:        handler,
			BatchSize:      config.Consumer.BatchSize,
			BlockDuration:  config.Consumer.BlockDuration,
			MaxConcurrency: config.Consumer.MaxConcurrency,
		})
	}()

	go func() {
		userSubscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:         streams.StreamUsers,
			ConsumerGroup:  config.Consumer.Group,
			ConsumerID:     config.Consumer.ConsumerID + "-users",
			Handler:        handler,
			BatchSize:      config.Consumer.BatchSize,
			BlockDuration:  config.Consumer.BlockDuration,
			MaxConcurrency: config.Consumer.MaxConcurrency,
		})
	}()
}
