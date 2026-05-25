//nolint:all // Test file
package redis_test

import (
	"context"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
	"github.com/tclavelloux/promy-event-bus/testutil"

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
		event := testutil.NewTestEvent("promotion.created", map[string]any{
			"promotion_id":   "promo-123",
			"promotion_name": "Patate douce",
			"distributor_id": "dist-456",
		})

		err := publisher.Publish(ctx, "events:test", event)
		assert.NoError(t, err)
	})

	t.Run("publishes different event types", func(t *testing.T) {
		promoEvent := testutil.NewTestEvent("promotion.created", map[string]any{
			"promotion_id":   "promo-456",
			"promotion_name": "Test Product",
			"distributor_id": "dist-123",
		})
		err := publisher.Publish(ctx, "events:promotions", promoEvent)
		assert.NoError(t, err)

		userEvent := testutil.NewTestEvent("user.registered", map[string]any{
			"user_id": "user-789",
			"email":   "test@example.com",
		})
		err = publisher.Publish(ctx, "events:users", userEvent)
		assert.NoError(t, err)

		productEvent := testutil.NewTestEvent("product.identified", map[string]any{
			"promotion_id": "promo-456",
			"product_id":   "prod-789",
			"confidence":   0.95,
		})
		err = publisher.Publish(ctx, "events:products", productEvent)
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
			testutil.NewTestEvent("promotion.created", map[string]any{"promotion_id": "promo-1"}),
			testutil.NewTestEvent("promotion.created", map[string]any{"promotion_id": "promo-2"}),
			testutil.NewTestEvent("promotion.created", map[string]any{"promotion_id": "promo-3"}),
		}

		err := publisher.PublishBatch(ctx, "events:test-batch", events)
		assert.NoError(t, err)
	})

	t.Run("handles empty batch", func(t *testing.T) {
		err := publisher.PublishBatch(ctx, "events:test-batch", []eventbus.Event{})
		assert.NoError(t, err)
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
