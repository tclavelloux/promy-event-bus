//nolint:all // Test file
package redis_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events/promotion"
	"github.com/tclavelloux/promy-event-bus/pkg/ptr"
	"github.com/tclavelloux/promy-event-bus/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriber_New(t *testing.T) {
	t.Run("creates subscriber with valid config", func(t *testing.T) {
		config := eventbus.Config{
			Redis: eventbus.RedisConfig{
				DSN: "redis://localhost:6379/1",
			},
			Consumer: eventbus.ConsumerConfig{
				Group:          "test-group",
				ConsumerID:     "test-consumer",
				BlockDuration:  1 * time.Second,
				BatchSize:      10,
				MaxConcurrency: 5,
			},
		}

		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		require.NotNil(t, subscriber)

		defer subscriber.Close()

		// Test health
		ctx := context.Background()
		err = subscriber.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("fails with invalid DSN", func(t *testing.T) {
		config := eventbus.Config{
			Redis: eventbus.RedisConfig{
				DSN: "invalid-dsn",
			},
		}

		subscriber, err := redis.NewSubscriber(config)
		assert.Error(t, err)
		assert.Nil(t, subscriber)
	})
}

func TestSubscriber_Subscribe(t *testing.T) {
	config := eventbus.Config{
		Redis: eventbus.RedisConfig{
			DSN: "redis://localhost:6379/1",
		},
		Consumer: eventbus.ConsumerConfig{
			Group:          "test-subscribe-group",
			ConsumerID:     "test-consumer-1",
			BlockDuration:  1 * time.Second,
			BatchSize:      10,
			MaxConcurrency: 5,
		},
	}

	t.Run("consumes published events", func(t *testing.T) {
		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		defer subscriber.Close()

		publisher, err := redis.NewPublisher(config.Redis)
		require.NoError(t, err)
		defer publisher.Close()

		// Channel to receive events
		received := make(chan eventbus.Event, 1)

		// Handler that captures events
		handler := func(ctx context.Context, event eventbus.Event) error {
			received <- event
			return nil
		}

		// Start subscriber in background
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go func() {
			subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
				Stream:         "events:test-subscribe",
				ConsumerGroup:  config.Consumer.Group,
				ConsumerID:     config.Consumer.ConsumerID,
				Handler:        handler,
				BatchSize:      config.Consumer.BatchSize,
				BlockDuration:  config.Consumer.BlockDuration,
				MaxConcurrency: config.Consumer.MaxConcurrency,
			})
		}()

		// Give subscriber time to start
		time.Sleep(100 * time.Millisecond)

		// Publish test event
		event := promotion.NewPromotionCreatedEvent(
			"promo-test",
			"Test Product",
			"dist-test",
			"leaflet-test",
			1,
			12.99,
			ptr.String("cat-test"),
			ptr.StringSlice([]string{"2025-11-06"}),
			ptr.String("https://example.com/test.jpg"),
			ptr.Float64(12.99),
		)

		err = publisher.Publish(context.Background(), "events:test-subscribe", event)
		require.NoError(t, err)

		// Wait for event
		select {
		case receivedEvent := <-received:
			assert.Equal(t, event.EventType(), receivedEvent.EventType())
			assert.Equal(t, event.EventID(), receivedEvent.EventID())
		case <-ctx.Done():
			t.Fatal("timeout waiting for event")
		}
	})

	t.Run("processes multiple events", func(t *testing.T) {
		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		defer subscriber.Close()

		publisher, err := redis.NewPublisher(config.Redis)
		require.NoError(t, err)
		defer publisher.Close()

		// Counter for received events
		var count atomic.Int32

		handler := func(ctx context.Context, event eventbus.Event) error {
			count.Add(1)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go func() {
			subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
				Stream:         "events:test-multiple",
				ConsumerGroup:  "test-multiple-group",
				ConsumerID:     "consumer-1",
				Handler:        handler,
				BatchSize:      10,
				BlockDuration:  1 * time.Second,
				MaxConcurrency: 5,
			})
		}()

		time.Sleep(100 * time.Millisecond)

		// Publish multiple events
		for i := 0; i < 5; i++ {
			event := promotion.NewPromotionCreatedEvent(
				"promo-"+string(rune(i)),
				"Product "+string(rune(i)),
				"dist-1",
				"leaflet-1",
				1,
				float64(i+1)*10.0, // Start at 10.0 to ensure price > 0
				ptr.String("cat-1"),
				ptr.StringSlice([]string{"2025-11-06"}),
				ptr.String("https://example.com/test.jpg"),
				ptr.Float64(float64(i+1)*10.0),
			)
			err = publisher.Publish(context.Background(), "events:test-multiple", event)
			require.NoError(t, err)
		}

		// Wait for all events to be processed
		time.Sleep(2 * time.Second)

		assert.Equal(t, int32(5), count.Load())
	})

	t.Run("handles handler errors with retry", func(t *testing.T) {
		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		defer subscriber.Close()

		publisher, err := redis.NewPublisher(config.Redis)
		require.NoError(t, err)
		defer publisher.Close()

		var attemptCount atomic.Int32

		handler := func(ctx context.Context, event eventbus.Event) error {
			count := attemptCount.Add(1)
			if count < 3 {
				return assert.AnError // Fail first 2 attempts
			}
			return nil // Succeed on 3rd attempt
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() {
			subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
				Stream:         "events:test-retry",
				ConsumerGroup:  "test-retry-group",
				ConsumerID:     "consumer-1",
				Handler:        handler,
				BatchSize:      10,
				BlockDuration:  1 * time.Second,
				MaxConcurrency: 1,
			})
		}()

		time.Sleep(100 * time.Millisecond)

		event := promotion.NewPromotionCreatedEvent(
			"promo-retry",
			"Retry Product",
			"dist-1",
			"leaflet-1",
			1,
			9.99,
			ptr.String("cat-1"),
			ptr.StringSlice([]string{"2025-11-06"}),
			ptr.String("https://example.com/retry.jpg"),
			ptr.Float64(9.99),
		)

		err = publisher.Publish(context.Background(), "events:test-retry", event)
		require.NoError(t, err)

		// Wait for retries
		time.Sleep(5 * time.Second)

		// Should have been called 3 times (initial + 2 retries)
		assert.GreaterOrEqual(t, attemptCount.Load(), int32(3))
	})
}

func TestSubscriber_Health(t *testing.T) {
	t.Run("returns healthy when connected", func(t *testing.T) {
		config := eventbus.Config{
			Redis: eventbus.RedisConfig{
				DSN: "redis://localhost:6379/1",
			},
		}

		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		defer subscriber.Close()

		ctx := context.Background()
		err = subscriber.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("returns unhealthy after close", func(t *testing.T) {
		config := eventbus.Config{
			Redis: eventbus.RedisConfig{
				DSN: "redis://localhost:6379/1",
			},
		}

		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)

		subscriber.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = subscriber.Health(ctx)
		assert.Error(t, err)
	})
}

func TestSubscriber_Close(t *testing.T) {
	t.Run("closes successfully", func(t *testing.T) {
		config := eventbus.Config{
			Redis: eventbus.RedisConfig{
				DSN: "redis://localhost:6379/1",
			},
		}

		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)

		err = subscriber.Close()
		assert.NoError(t, err)
	})
}
