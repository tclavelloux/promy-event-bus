//nolint:all // Test file
package redis_test

import (
	"context"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
	"github.com/tclavelloux/promy-event-bus/events/product"
	"github.com/tclavelloux/promy-event-bus/events/promotion"
	"github.com/tclavelloux/promy-event-bus/events/user"
	"github.com/tclavelloux/promy-event-bus/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisher_New(t *testing.T) {
	t.Run("creates publisher with valid config", func(t *testing.T) {
		config := eventbus.RedisConfig{
			DSN: "redis://localhost:6379/1",
		}

		publisher, err := redis.NewPublisher(config)
		require.NoError(t, err)
		require.NotNil(t, publisher)

		defer publisher.Close()

		// Test health
		ctx := context.Background()
		err = publisher.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("fails with invalid DSN", func(t *testing.T) {
		config := eventbus.RedisConfig{
			DSN: "invalid-dsn",
		}

		publisher, err := redis.NewPublisher(config)
		assert.Error(t, err)
		assert.Nil(t, publisher)
	})

	t.Run("fails with unreachable Redis", func(t *testing.T) {
		config := eventbus.RedisConfig{
			DSN: "redis://localhost:9999/0",
		}

		publisher, err := redis.NewPublisher(config)
		assert.Error(t, err)
		assert.Nil(t, publisher)
	})
}

func TestPublisher_Publish(t *testing.T) {
	config := eventbus.RedisConfig{
		DSN: "redis://localhost:6379/1",
	}

	publisher, err := redis.NewPublisher(config)
	require.NoError(t, err)
	defer publisher.Close()

	ctx := context.Background()

	t.Run("publishes event successfully", func(t *testing.T) {
		event := promotion.NewPromotionCreatedEvent(
			"promo-123",
			"Patate douce",
			"dist-456",
			"cat-789",
			[]string{"2025-11-06"},
			9.99,
			"https://example.com/image.jpg",
		)

		err := publisher.Publish(ctx, "events:test", event)
		assert.NoError(t, err)
	})

	t.Run("validates event before publishing", func(t *testing.T) {
		// Create invalid event (missing required fields)
		event := &promotion.PromotionCreatedEvent{
			BaseEvent: eventbus.NewBaseEvent(events.EventPromotionCreated, "test"),
			// Missing PromotionID, ProductName, DistributorID
		}

		err := publisher.Publish(ctx, "events:test", event)
		assert.Error(t, err)
		assert.ErrorIs(t, err, eventbus.ErrInvalidEvent)
	})

	t.Run("publishes different event types", func(t *testing.T) {
		// Promotion event
		promoEvent := promotion.NewPromotionCreatedEvent(
			"promo-456",
			"Test Product",
			"dist-123",
			"cat-456",
			[]string{"2025-11-07"},
			15.99,
			"https://example.com/test.jpg",
		)
		err := publisher.Publish(ctx, events.StreamPromotions, promoEvent)
		assert.NoError(t, err)

		// User event
		userEvent := user.NewUserRegisteredEvent("user-789", "test@example.com")
		err = publisher.Publish(ctx, events.StreamUsers, userEvent)
		assert.NoError(t, err)

		// Product event
		productEvent := product.NewProductIdentifiedEvent(
			"promo-456",
			"prod-789",
			"vegetables",
			"cat-123",
			"BrandX",
			0.95,
		)
		err = publisher.Publish(ctx, events.StreamProducts, productEvent)
		assert.NoError(t, err)
	})
}

func TestPublisher_PublishBatch(t *testing.T) {
	config := eventbus.RedisConfig{
		DSN: "redis://localhost:6379/1",
	}

	publisher, err := redis.NewPublisher(config)
	require.NoError(t, err)
	defer publisher.Close()

	ctx := context.Background()

	t.Run("publishes batch successfully", func(t *testing.T) {
		events := []eventbus.Event{
			promotion.NewPromotionCreatedEvent(
				"promo-1",
				"Product 1",
				"dist-1",
				"cat-1",
				[]string{"2025-11-06"},
				10.00,
				"https://example.com/1.jpg",
			),
			promotion.NewPromotionCreatedEvent(
				"promo-2",
				"Product 2",
				"dist-2",
				"cat-2",
				[]string{"2025-11-06"},
				20.00,
				"https://example.com/2.jpg",
			),
			promotion.NewPromotionCreatedEvent(
				"promo-3",
				"Product 3",
				"dist-3",
				"cat-3",
				[]string{"2025-11-06"},
				30.00,
				"https://example.com/3.jpg",
			),
		}

		err := publisher.PublishBatch(ctx, "events:test-batch", events)
		assert.NoError(t, err)
	})

	t.Run("handles empty batch", func(t *testing.T) {
		err := publisher.PublishBatch(ctx, "events:test-batch", []eventbus.Event{})
		assert.NoError(t, err)
	})

	t.Run("fails with invalid event in batch", func(t *testing.T) {
		events := []eventbus.Event{
			promotion.NewPromotionCreatedEvent(
				"promo-1",
				"Product 1",
				"dist-1",
				"cat-1",
				[]string{"2025-11-06"},
				10.00,
				"https://example.com/1.jpg",
			),
			&promotion.PromotionCreatedEvent{
				BaseEvent: eventbus.NewBaseEvent(events.EventPromotionCreated, "test"),
				// Invalid: missing required fields
			},
		}

		err := publisher.PublishBatch(ctx, "events:test-batch", events)
		assert.Error(t, err)
		assert.ErrorIs(t, err, eventbus.ErrInvalidEvent)
	})
}

func TestPublisher_Health(t *testing.T) {
	t.Run("returns healthy when connected", func(t *testing.T) {
		config := eventbus.RedisConfig{
			DSN: "redis://localhost:6379/1",
		}

		publisher, err := redis.NewPublisher(config)
		require.NoError(t, err)
		defer publisher.Close()

		ctx := context.Background()
		err = publisher.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("returns unhealthy after close", func(t *testing.T) {
		config := eventbus.RedisConfig{
			DSN: "redis://localhost:6379/1",
		}

		publisher, err := redis.NewPublisher(config)
		require.NoError(t, err)

		publisher.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = publisher.Health(ctx)
		assert.Error(t, err)
	})
}

func TestPublisher_Close(t *testing.T) {
	t.Run("closes successfully", func(t *testing.T) {
		config := eventbus.RedisConfig{
			DSN: "redis://localhost:6379/1",
		}

		publisher, err := redis.NewPublisher(config)
		require.NoError(t, err)

		err = publisher.Close()
		assert.NoError(t, err)
	})
}
