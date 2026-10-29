//nolint:all // Test file
package redis_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDLQDown = errors.New("dlq down")

// failingPublisher is an EventPublisher whose Publish always fails.
type failingPublisher struct{}

func (failingPublisher) Publish(context.Context, string, eventbus.Event) error { return errDLQDown }
func (failingPublisher) PublishBatch(context.Context, string, []eventbus.Event) error {
	return errDLQDown
}
func (failingPublisher) Close() error                 { return nil }
func (failingPublisher) Health(context.Context) error { return nil }

// errCollector records SubscriberErrors reported through OnError.
type errCollector struct {
	mu   sync.Mutex
	errs []eventbus.SubscriberError
}

func (c *errCollector) record(_ context.Context, e eventbus.SubscriberError) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errs = append(c.errs, e)
}

func (c *errCollector) snapshot() []eventbus.SubscriberError {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]eventbus.SubscriberError(nil), c.errs...)
}

// waitFor polls until n errors were collected or the context ends.
func (c *errCollector) waitFor(t *testing.T, ctx context.Context, n int) []eventbus.SubscriberError {
	t.Helper()

	for {
		if got := c.snapshot(); len(got) >= n {
			return got
		}

		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for %d reported errors, got %d", n, len(c.snapshot()))
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// onErrorFixture is a single-message subscription against a unique stream.
type onErrorFixture struct {
	client    *goredis.Client
	sub       *redis.Subscriber
	stream    string
	group     string
	consumer  string
	eventID   string
	messageID string
}

func newOnErrorFixture(t *testing.T, attempt int) *onErrorFixture {
	t.Helper()

	sub, err := redis.NewSubscriber(eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sub.Close() })

	f := &onErrorFixture{
		client:   testClient(t),
		sub:      sub,
		stream:   uniqueStream(t, "onerror"),
		group:    "onerror-group",
		consumer: "onerror-consumer",
		eventID:  uuid.NewString(),
	}

	metadataJSON, err := json.Marshal(map[string]any{
		"id":        f.eventID,
		"type":      "user.registered",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0",
		"attempt":   attempt,
	})
	require.NoError(t, err)

	f.messageID, err = f.client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: f.stream,
		Values: map[string]any{"metadata": string(metadataJSON), "payload": `{"user_id":"u-1"}`},
	}).Result()
	require.NoError(t, err)

	return f
}

// breakStream replaces the stream key with a string so later XADD/XACK return WRONGTYPE.
func (f *onErrorFixture) breakStream(ctx context.Context) {
	f.client.Del(ctx, f.stream)
	f.client.Set(ctx, f.stream, "not-a-stream", 0)
}

func (f *onErrorFixture) subscribe(ctx context.Context, cfg eventbus.SubscriptionConfig) <-chan error {
	cfg.Stream, cfg.ConsumerGroup, cfg.ConsumerID = f.stream, f.group, f.consumer
	done := make(chan error, 1)

	go func() { done <- f.sub.Subscribe(ctx, cfg) }()

	return done
}

func (f *onErrorFixture) assertCommon(t *testing.T, e eventbus.SubscriberError, stage eventbus.SubscriberErrorStage, attempt int) {
	t.Helper()

	assert.Equal(t, stage, e.Stage)
	assert.Equal(t, f.stream, e.Stream)
	assert.Equal(t, f.group, e.ConsumerGroup)
	assert.Equal(t, f.consumer, e.ConsumerID)
	assert.Equal(t, f.messageID, e.MessageID)
	assert.Equal(t, f.eventID, e.EventID)
	assert.Equal(t, "user.registered", e.EventType)
	assert.Equal(t, attempt, e.Attempt)
}

func TestSubscriber_OnError_DLQPublishFails(t *testing.T) {
	f := newOnErrorFixture(t, 3)
	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	f.subscribe(ctx, eventbus.SubscriptionConfig{
		DLQPublisher: failingPublisher{},
		OnError:      col.record,
		Handler:      func(context.Context, eventbus.Event) error { return assert.AnError },
	})

	col.waitFor(t, ctx, 1)
	<-ctx.Done()

	errs := col.snapshot()
	require.Len(t, errs, 1)
	f.assertCommon(t, errs[0], eventbus.StageDLQPublish, 3)
	assert.ErrorIs(t, errs[0].Err, errDLQDown)
	assert.ErrorIs(t, errs[0], errDLQDown)

	pending, err := f.client.XPending(context.Background(), f.stream, f.group).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count)
}

func TestSubscriber_OnError_RequeueAndAckFail(t *testing.T) {
	f := newOnErrorFixture(t, 1)
	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := f.subscribe(ctx, eventbus.SubscriptionConfig{
		OnError: col.record,
		Handler: func(ctx context.Context, _ eventbus.Event) error {
			f.breakStream(ctx)

			return assert.AnError
		},
	})

	errs := col.waitFor(t, ctx, 2)
	require.Len(t, errs, 2)

	f.assertCommon(t, errs[0], eventbus.StageRetryRequeue, 1)
	assert.ErrorContains(t, errs[0].Err, "WRONGTYPE")
	f.assertCommon(t, errs[1], eventbus.StageAck, 1)
	assert.ErrorContains(t, errs[1].Err, "WRONGTYPE")

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-ctx.Done():
		t.Fatal("Subscribe did not return")
	}
}

func TestSubscriber_OnError_AckFailsOnSuccess(t *testing.T) {
	f := newOnErrorFixture(t, 1)
	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	f.subscribe(ctx, eventbus.SubscriptionConfig{
		OnError: col.record,
		Handler: func(ctx context.Context, _ eventbus.Event) error {
			f.breakStream(ctx)

			return nil
		},
	})

	errs := col.waitFor(t, ctx, 1)
	cancel()

	require.Len(t, errs, 1)
	f.assertCommon(t, errs[0], eventbus.StageAck, 1)
	assert.ErrorContains(t, errs[0].Err, "WRONGTYPE")
}

func TestSubscriber_OnError_NilCallbackDoesNotPanic(t *testing.T) {
	f := newOnErrorFixture(t, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := f.subscribe(ctx, eventbus.SubscriptionConfig{
		Handler: func(ctx context.Context, _ eventbus.Event) error {
			f.breakStream(ctx)

			return assert.AnError
		},
	})

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-ctx.Done():
		t.Fatal("Subscribe did not return")
	}
}

func TestSubscriber_OnError_PanicIsRecovered(t *testing.T) {
	f := newOnErrorFixture(t, 1)
	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := f.subscribe(ctx, eventbus.SubscriptionConfig{
		OnError: func(ctx context.Context, e eventbus.SubscriberError) {
			col.record(ctx, e)
			panic("observer bug")
		},
		Handler: func(ctx context.Context, _ eventbus.Event) error {
			f.breakStream(ctx)

			return assert.AnError
		},
	})

	errs := col.waitFor(t, ctx, 2)
	require.Len(t, errs, 2)
	assert.Equal(t, eventbus.StageRetryRequeue, errs[0].Stage)
	assert.Equal(t, eventbus.StageAck, errs[1].Stage)

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-ctx.Done():
		t.Fatal("Subscribe did not return")
	}
}

func TestSubscriber_OnError_MetadataMissing(t *testing.T) {
	sub, err := redis.NewSubscriber(eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}})
	require.NoError(t, err)
	defer sub.Close()

	client := testClient(t)
	stream := uniqueStream(t, "onerror-missing")

	msgID, err := client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: stream,
		Values: map[string]any{"payload": `{"foo":"bar"}`},
	}).Result()
	require.NoError(t, err)

	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = sub.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream: stream, ConsumerGroup: "g", ConsumerID: "c",
			OnError: col.record,
			Handler: func(context.Context, eventbus.Event) error { return nil },
		})
	}()

	errs := col.waitFor(t, ctx, 1)
	cancel()

	require.Len(t, errs, 1)
	assert.Equal(t, eventbus.StageMetadataUnmarshal, errs[0].Stage)
	assert.Equal(t, stream, errs[0].Stream)
	assert.Equal(t, "g", errs[0].ConsumerGroup)
	assert.Equal(t, "c", errs[0].ConsumerID)
	assert.Equal(t, msgID, errs[0].MessageID)
	assert.Empty(t, errs[0].EventID)
	assert.Empty(t, errs[0].EventType)
	assert.Zero(t, errs[0].Attempt)
	assert.ErrorContains(t, errs[0].Err, "metadata field missing")

	pending, err := client.XPending(context.Background(), stream, "g").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count)
}

func TestSubscriber_OnError_MetadataMalformed(t *testing.T) {
	sub, err := redis.NewSubscriber(eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}})
	require.NoError(t, err)
	defer sub.Close()

	client := testClient(t)
	stream := uniqueStream(t, "onerror-malformed")

	msgID, err := client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: stream,
		Values: map[string]any{"metadata": "not-json", "payload": `{"foo":"bar"}`},
	}).Result()
	require.NoError(t, err)

	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = sub.Subscribe(ctx, eventbus.SubscriptionConfig{
			Stream: stream, ConsumerGroup: "g", ConsumerID: "c",
			OnError: col.record,
			Handler: func(context.Context, eventbus.Event) error { return nil },
		})
	}()

	errs := col.waitFor(t, ctx, 1)
	cancel()

	require.Len(t, errs, 1)
	assert.Equal(t, eventbus.StageMetadataUnmarshal, errs[0].Stage)
	assert.Equal(t, msgID, errs[0].MessageID)
	assert.Empty(t, errs[0].EventID)
	assert.Zero(t, errs[0].Attempt)

	var syntaxErr *json.SyntaxError
	assert.ErrorAs(t, errs[0], &syntaxErr)

	pending, err := client.XPending(context.Background(), stream, "g").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count)
}
