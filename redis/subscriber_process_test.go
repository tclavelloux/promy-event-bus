//nolint:all // Test file
package redis_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
	"github.com/tclavelloux/promy-event-bus/streams"
	"github.com/tclavelloux/promy-event-bus/testutil"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubscriber_AppliesAllOptions(t *testing.T) {
	config := eventbus.Config{
		Redis: eventbus.RedisConfig{
			DSN:             testDSN(t),
			PoolSize:        7,
			MaxRetries:      4,
			MinRetryBackoff: 5 * time.Millisecond,
			MaxRetryBackoff: 200 * time.Millisecond,
			DialTimeout:     2 * time.Second,
			ReadTimeout:     time.Second,
			WriteTimeout:    time.Second,
		},
	}

	subscriber, err := redis.NewSubscriber(config)
	require.NoError(t, err)
	defer subscriber.Close()

	assert.NoError(t, subscriber.Health(t.Context()))
}

func TestNewSubscriber_UnreachableRedis(t *testing.T) {
	config := eventbus.Config{
		Redis: eventbus.RedisConfig{DSN: "redis://localhost:9999/0"},
	}

	subscriber, err := redis.NewSubscriber(config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
	assert.Nil(t, subscriber)
}

func TestSubscriber_Subscribe_AppliesDefaults(t *testing.T) {
	config := eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}}

	subscriber, err := redis.NewSubscriber(config)
	require.NoError(t, err)
	defer subscriber.Close()

	publisher, err := redis.NewPublisher(config.Redis)
	require.NoError(t, err)
	defer publisher.Close()

	stream := uniqueStream(t, "defaults")
	received := make(chan eventbus.Event, 1)

	handler := func(ctx context.Context, event eventbus.Event) error {
		received <- event

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	subErrCh := make(chan error, 1)

	go func() {
		subErrCh <- subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:        stream,
			ConsumerGroup: "defaults-group",
			ConsumerID:    "defaults-consumer",
			Handler:       handler,
			// MaxConcurrency, BatchSize, BlockDuration all left zero.
		})
	}()

	time.Sleep(100 * time.Millisecond)

	event := testutil.NewTestEvent("user.registered", map[string]any{"user_id": "u-defaults"})
	require.NoError(t, publisher.Publish(context.Background(), stream, event))

	select {
	case receivedEvent := <-received:
		assert.Equal(t, event.EventID(), receivedEvent.EventID())
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}

	select {
	case err := <-subErrCh:
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(4 * time.Second):
		t.Fatal("Subscribe did not return after context cancellation")
	}
}

func TestSubscriber_Subscribe_ConsumerGroupCreateFails(t *testing.T) {
	config := eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}}

	subscriber, err := redis.NewSubscriber(config)
	require.NoError(t, err)
	defer subscriber.Close()

	stream := uniqueStream(t, "wrongtype")

	client := testClient(t)
	require.NoError(t, client.Set(context.Background(), stream, "not-a-stream", 0).Err())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
		Stream:        stream,
		ConsumerGroup: "wrongtype-group",
		ConsumerID:    "wrongtype-consumer",
		Handler: func(ctx context.Context, event eventbus.Event) error {
			return nil
		},
	})

	assert.ErrorIs(t, err, eventbus.ErrSubscriptionFailed)
}

func TestSubscriber_ProcessMessage_MissingMetadata(t *testing.T) {
	config := eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}}

	subscriber, err := redis.NewSubscriber(config)
	require.NoError(t, err)
	defer subscriber.Close()

	stream := uniqueStream(t, "missing-meta")
	group := "missing-meta-group"

	client := testClient(t)
	_, err = client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: stream,
		Values: map[string]any{"payload": `{"foo":"bar"}`},
	}).Result()
	require.NoError(t, err)

	var handlerCalled atomic.Bool

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:        stream,
			ConsumerGroup: group,
			ConsumerID:    "consumer-1",
			Handler: func(ctx context.Context, event eventbus.Event) error {
				handlerCalled.Store(true)

				return nil
			},
		})
	}()

	<-ctx.Done()

	assert.False(t, handlerCalled.Load())

	pending, err := client.XPending(context.Background(), stream, group).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count)
}

func TestSubscriber_ProcessMessage_MalformedMetadata(t *testing.T) {
	config := eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}}

	subscriber, err := redis.NewSubscriber(config)
	require.NoError(t, err)
	defer subscriber.Close()

	stream := uniqueStream(t, "malformed-meta")
	group := "malformed-meta-group"

	client := testClient(t)
	_, err = client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: stream,
		Values: map[string]any{"metadata": "not-json", "payload": `{"foo":"bar"}`},
	}).Result()
	require.NoError(t, err)

	var handlerCalled atomic.Bool

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:        stream,
			ConsumerGroup: group,
			ConsumerID:    "consumer-1",
			Handler: func(ctx context.Context, event eventbus.Event) error {
				handlerCalled.Store(true)

				return nil
			},
		})
	}()

	<-ctx.Done()

	assert.False(t, handlerCalled.Load())

	pending, err := client.XPending(context.Background(), stream, group).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count)
}

func TestSubscriber_ProcessMessage_RoutesToDLQAfterExhaustion(t *testing.T) {
	config := eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}}

	subscriber, err := redis.NewSubscriber(config)
	require.NoError(t, err)
	defer subscriber.Close()

	dlqPublisher, err := redis.NewPublisher(config.Redis)
	require.NoError(t, err)
	defer dlqPublisher.Close()

	stream := uniqueStream(t, "dlq-exhausted")
	group := "dlq-exhausted-group"

	client := testClient(t)

	// events:dlq is a shared, unprefixed stream — make sure it starts empty
	// for this test's assertion (no other test in this package writes to it).
	client.Del(context.Background(), streams.StreamDLQ)

	eventID := uuid.NewString()
	timestamp := time.Now().UTC().Format(time.RFC3339)
	metadata := map[string]any{
		"id":        eventID,
		"type":      "user.registered",
		"timestamp": timestamp,
		"version":   "1.0",
		"attempt":   3,
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	_, err = client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"metadata": string(metadataJSON),
			"payload":  `{"user_id":"u-1"}`,
		},
	}).Result()
	require.NoError(t, err)

	var handlerEvent eventbus.Event

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:        stream,
			ConsumerGroup: group,
			ConsumerID:    "consumer-1",
			DLQPublisher:  dlqPublisher,
			DLQService:    "promy-test",
			Handler: func(ctx context.Context, event eventbus.Event) error {
				handlerEvent = event

				return assert.AnError
			},
		})
	}()

	// Poll until the DLQ entry shows up, bounded by the parent context.
	var dlqLen int64
	for {
		dlqLen, err = client.XLen(context.Background(), streams.StreamDLQ).Result()
		require.NoError(t, err)

		if dlqLen > 0 {
			break
		}

		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for DLQ entry")
		case <-time.After(50 * time.Millisecond):
		}
	}

	require.NotNil(t, handlerEvent)
	assert.Equal(t, timestamp, handlerEvent.EventTime().Format(time.RFC3339))
	assert.NoError(t, handlerEvent.Validate())

	assert.Equal(t, int64(1), dlqLen)

	entries, err := client.XRange(context.Background(), streams.StreamDLQ, "-", "+").Result()
	require.NoError(t, err)
	require.Len(t, entries, 1)

	payloadStr, ok := entries[0].Values["payload"].(string)
	require.True(t, ok)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(payloadStr), &decoded))

	assert.Equal(t, stream, decoded["original_stream"])
	assert.Equal(t, "user.registered", decoded["original_event_type"])
	assert.Equal(t, "promy-test", decoded["failed_service"])
	assert.Equal(t, float64(3), decoded["attempts_exhausted"])
}

func TestSubscriber_ProcessMessage_DropsWhenNoDLQPublisher(t *testing.T) {
	config := eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}}

	subscriber, err := redis.NewSubscriber(config)
	require.NoError(t, err)
	defer subscriber.Close()

	stream := uniqueStream(t, "dlq-dropped")
	group := "dlq-dropped-group"

	client := testClient(t)

	dlqLenBefore, err := client.XLen(context.Background(), streams.StreamDLQ).Result()
	require.NoError(t, err)

	metadata := map[string]any{
		"id":        uuid.NewString(),
		"type":      "user.registered",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0",
		"attempt":   3,
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	_, err = client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"metadata": string(metadataJSON),
			"payload":  `{"user_id":"u-1"}`,
		},
	}).Result()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream:        stream,
			ConsumerGroup: group,
			ConsumerID:    "consumer-1",
			Handler: func(ctx context.Context, event eventbus.Event) error {
				return assert.AnError
			},
		})
	}()

	<-ctx.Done()

	pending, err := client.XPending(context.Background(), stream, group).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count)

	dlqLenAfter, err := client.XLen(context.Background(), streams.StreamDLQ).Result()
	require.NoError(t, err)
	assert.Equal(t, dlqLenBefore, dlqLenAfter)
}
