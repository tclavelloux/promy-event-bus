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
	"github.com/tclavelloux/promy-event-bus/pkg/ptr"
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
			"leaflet-123",
			1,
			9.99,
			ptr.String("cat-789"),
			ptr.StringSlice([]string{"2025-11-06"}),
			ptr.String("https://example.com/image.jpg"),
			ptr.Float64(11.99),
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
			"leaflet-456",
			1,
			15.99,
			ptr.String("cat-456"),
			ptr.StringSlice([]string{"2025-11-07"}),
			ptr.String("https://example.com/test.jpg"),
			ptr.Float64(15.99),
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
			ptr.String("BrandX"),
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
				"leaflet-1",
				1,
				10.00,
				ptr.String("cat-1"),
				ptr.StringSlice([]string{"2025-11-06"}),
				ptr.String("https://example.com/1.jpg"),
				ptr.Float64(10.00),
			),
			promotion.NewPromotionCreatedEvent(
				"promo-2",
				"Product 2",
				"dist-2",
				"leaflet-2",
				2,
				20.00,
				ptr.String("cat-2"),
				ptr.StringSlice([]string{"2025-11-06"}),
				ptr.String("https://example.com/2.jpg"),
				ptr.Float64(20.00),
			),
			promotion.NewPromotionCreatedEvent(
				"promo-3",
				"Product 3",
				"dist-3",
				"leaflet-3",
				3,
				30.00,
				ptr.String("cat-3"),
				ptr.StringSlice([]string{"2025-11-06"}),
				ptr.String("https://example.com/3.jpg"),
				ptr.Float64(30.00),
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
				"leaflet-1",
				1,
				10.00,
				ptr.String("cat-1"),
				ptr.StringSlice([]string{"2025-11-06"}),
				ptr.String("https://example.com/1.jpg"),
				ptr.Float64(10.00),
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
