//nolint:all // Test file
package redis_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
	"github.com/tclavelloux/promy-event-bus/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_PublishAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := eventbus.Config{
		Redis: eventbus.RedisConfig{
			DSN: "redis://localhost:6379/1",
		},
		Consumer: eventbus.ConsumerConfig{
			Group:          "integration-test-group",
			ConsumerID:     "consumer-1",
			BlockDuration:  1 * time.Second,
			BatchSize:      10,
			MaxConcurrency: 5,
		},
	}

	t.Run("end-to-end event flow", func(t *testing.T) {
		publisher, err := redis.NewPublisher(config.Redis)
		require.NoError(t, err)
		defer publisher.Close()

		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		defer subscriber.Close()

		receivedEvents := make(map[string]bool)
		var mu sync.Mutex

		handler := func(ctx context.Context, event eventbus.Event) error {
			mu.Lock()
			receivedEvents[event.EventID()] = true
			mu.Unlock()
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() {
			subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
				Stream:         "events:integration-test",
				ConsumerGroup:  config.Consumer.Group,
				ConsumerID:     config.Consumer.ConsumerID,
				Handler:        handler,
				BatchSize:      config.Consumer.BatchSize,
				BlockDuration:  config.Consumer.BlockDuration,
				MaxConcurrency: config.Consumer.MaxConcurrency,
			})
		}()

		time.Sleep(100 * time.Millisecond)

		promoEvent := testutil.NewTestEvent("promotion.created", map[string]any{
			"promotion_id":   "integration-promo-1",
			"promotion_name": "Integration Product",
			"distributor_id": "dist-1",
		})
		err = publisher.Publish(context.Background(), "events:integration-test", promoEvent)
		require.NoError(t, err)

		userEvent := testutil.NewTestEvent("user.registered", map[string]any{
			"user_id": "user-integration-1",
			"email":   "integration@example.com",
		})
		err = publisher.Publish(context.Background(), "events:integration-test", userEvent)
		require.NoError(t, err)

		productEvent := testutil.NewTestEvent("product.identified", map[string]any{
			"promotion_id": "integration-promo-1",
			"product_id":   "prod-integration-1",
			"confidence":   0.92,
		})
		err = publisher.Publish(context.Background(), "events:integration-test", productEvent)
		require.NoError(t, err)

		time.Sleep(3 * time.Second)

		mu.Lock()
		assert.True(t, receivedEvents[promoEvent.EventID()], "promotion event should be received")
		assert.True(t, receivedEvents[userEvent.EventID()], "user event should be received")
		assert.True(t, receivedEvents[productEvent.EventID()], "product event should be received")
		mu.Unlock()
	})

	t.Run("multiple consumers in same group share workload", func(t *testing.T) {
		publisher, err := redis.NewPublisher(config.Redis)
		require.NoError(t, err)
		defer publisher.Close()

		subscriber1, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		defer subscriber1.Close()

		config2 := config
		config2.Consumer.ConsumerID = "consumer-2"
		subscriber2, err := redis.NewSubscriber(config2)
		require.NoError(t, err)
		defer subscriber2.Close()

		var count1, count2 atomic.Int32

		handler1 := func(ctx context.Context, event eventbus.Event) error {
			count1.Add(1)
			return nil
		}

		handler2 := func(ctx context.Context, event eventbus.Event) error {
			count2.Add(1)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() {
			subscriber1.Subscribe(ctx, eventbus.SubscriptionConfig{
				Stream:         "events:load-balance-test",
				ConsumerGroup:  "load-balance-group",
				ConsumerID:     "consumer-1",
				Handler:        handler1,
				BatchSize:      1,
				BlockDuration:  1 * time.Second,
				MaxConcurrency: 1,
			})
		}()

		go func() {
			subscriber2.Subscribe(ctx, eventbus.SubscriptionConfig{
				Stream:         "events:load-balance-test",
				ConsumerGroup:  "load-balance-group",
				ConsumerID:     "consumer-2",
				Handler:        handler2,
				BatchSize:      1,
				BlockDuration:  1 * time.Second,
				MaxConcurrency: 1,
			})
		}()

		time.Sleep(100 * time.Millisecond)

		for i := 0; i < 10; i++ {
			event := testutil.NewTestEvent("promotion.created", map[string]any{
				"promotion_id": "load-balance-" + string(rune('a'+i)),
			})
			err = publisher.Publish(context.Background(), "events:load-balance-test", event)
			require.NoError(t, err)
		}

		time.Sleep(5 * time.Second)

		total := count1.Load() + count2.Load()
		assert.Equal(t, int32(10), total, "all events should be processed")
		assert.Greater(t, count1.Load(), int32(0), "consumer 1 should process some events")
		assert.Greater(t, count2.Load(), int32(0), "consumer 2 should process some events")
	})

	t.Run("batch publishing and consuming", func(t *testing.T) {
		publisher, err := redis.NewPublisher(config.Redis)
		require.NoError(t, err)
		defer publisher.Close()

		subscriber, err := redis.NewSubscriber(config)
		require.NoError(t, err)
		defer subscriber.Close()

		var count atomic.Int32

		handler := func(ctx context.Context, event eventbus.Event) error {
			count.Add(1)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() {
			subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
				Stream:         "events:batch-test",
				ConsumerGroup:  "batch-test-group",
				ConsumerID:     "consumer-1",
				Handler:        handler,
				BatchSize:      50,
				BlockDuration:  1 * time.Second,
				MaxConcurrency: 10,
			})
		}()

		time.Sleep(100 * time.Millisecond)

		batchEvents := make([]eventbus.Event, 20)
		for i := 0; i < 20; i++ {
			batchEvents[i] = testutil.NewTestEvent("promotion.created", map[string]any{
				"promotion_id": "batch-" + string(rune('a'+i)),
			})
		}

		err = publisher.PublishBatch(context.Background(), "events:batch-test", batchEvents)
		require.NoError(t, err)

		time.Sleep(3 * time.Second)
		assert.Equal(t, int32(20), count.Load(), "all batch events should be processed")
	})
}
