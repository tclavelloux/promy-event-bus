//nolint:all // Test file
package redis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
	"github.com/tclavelloux/promy-event-bus/testutil"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// lockedBuffer is a goroutine-safe log sink.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

// lines returns the JSON log lines written so far.
func (b *lockedBuffer) lines(t *testing.T) []map[string]any {
	t.Helper()

	b.mu.Lock()
	defer b.mu.Unlock()

	var out []map[string]any

	for _, l := range strings.Split(strings.TrimSpace(b.buf.String()), "\n") {
		if l == "" {
			continue
		}

		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(l), &m))
		out = append(out, m)
	}

	return out
}

// waitLines polls until n log lines exist or the context ends.
func (b *lockedBuffer) waitLines(t *testing.T, ctx context.Context, n int) []map[string]any {
	t.Helper()

	for {
		if got := b.lines(t); len(got) >= n {
			return got
		}

		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for %d log lines, got %d", n, len(b.lines(t)))
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// newTracing returns an in-memory exporter and a provider feeding it synchronously.
func newTracing(t *testing.T) (*tracetest.InMemoryExporter, *sdktrace.TracerProvider) {
	t.Helper()

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	return exp, tp
}

// waitSpans polls until n finished spans named name exist and returns them.
func waitSpans(t *testing.T, ctx context.Context, exp *tracetest.InMemoryExporter, name string, n int) []tracetest.SpanStub {
	t.Helper()

	for {
		var got []tracetest.SpanStub

		for _, s := range exp.GetSpans() {
			if s.Name == name {
				got = append(got, s)
			}
		}

		if len(got) >= n {
			return got
		}

		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for %d %q spans, got %d", n, name, len(got))
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func attrValue(s tracetest.SpanStub, key string) (attribute.Value, bool) {
	for _, kv := range s.Attributes {
		if string(kv.Key) == key {
			return kv.Value, true
		}
	}

	return attribute.Value{}, false
}

func requireAttr(t *testing.T, s tracetest.SpanStub, key string) attribute.Value {
	t.Helper()

	v, ok := attrValue(s, key)
	require.True(t, ok, "attribute %s missing", key)

	return v
}

func hasExceptionEvent(s tracetest.SpanStub) bool {
	for _, e := range s.Events {
		if e.Name == "exception" {
			return true
		}
	}

	return false
}

// tracingFixture wires a publisher and a subscriber on a unique stream with one tracer provider.
type tracingFixture struct {
	exp    *tracetest.InMemoryExporter
	tp     *sdktrace.TracerProvider
	pub    *redis.Publisher
	sub    *redis.Subscriber
	client *goredis.Client
	stream string
}

func newTracingFixture(t *testing.T, subOpts ...redis.Option) *tracingFixture {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	exp, tp := newTracing(t)
	dsn := testDSN(t)

	pub, err := redis.NewPublisher(eventbus.RedisConfig{DSN: dsn}, redis.WithTracerProvider(tp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = pub.Close() })

	sub, err := redis.NewSubscriber(
		eventbus.Config{Redis: eventbus.RedisConfig{DSN: dsn}},
		append([]redis.Option{redis.WithTracerProvider(tp)}, subOpts...)...,
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sub.Close() })

	return &tracingFixture{exp: exp, tp: tp, pub: pub, sub: sub, client: testClient(t), stream: uniqueStream(t, "trace")}
}

// subscribe runs Subscribe until the returned stop function is called.
func (f *tracingFixture) subscribe(ctx context.Context, cfg eventbus.SubscriptionConfig) (stop func()) {
	cfg.Stream, cfg.ConsumerGroup, cfg.ConsumerID = f.stream, "trace-group", "trace-consumer"

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		_ = f.sub.Subscribe(ctx, cfg)
	}()

	return func() {
		cancel()
		<-done
	}
}

func (f *tracingFixture) metadata(t *testing.T, msg goredis.XMessage) map[string]any {
	t.Helper()

	var md map[string]any
	require.NoError(t, json.Unmarshal([]byte(msg.Values["metadata"].(string)), &md))

	return md
}

func TestTracing_RoundTrip(t *testing.T) {
	f := newTracingFixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	parentCtx, parent := f.tp.Tracer("test").Start(ctx, "parent")
	event := testutil.NewTestEvent("promotion.created", map[string]any{"promotion_id": "p-1"})
	require.NoError(t, f.pub.Publish(parentCtx, f.stream, event))
	parent.End()

	handlerSC := make(chan trace.SpanContext, 1)
	stop := f.subscribe(ctx, eventbus.SubscriptionConfig{
		Handler: func(ctx context.Context, _ eventbus.Event) error {
			handlerSC <- trace.SpanContextFromContext(ctx)

			return nil
		},
	})
	defer stop()

	producer := waitSpans(t, ctx, f.exp, f.stream+".publish", 1)[0]
	consumer := waitSpans(t, ctx, f.exp, f.stream+".process", 1)[0]

	var hsc trace.SpanContext
	select {
	case hsc = <-handlerSC:
	case <-ctx.Done():
		t.Fatal("handler not called")
	}

	traceID := parent.SpanContext().TraceID()
	assert.Equal(t, traceID, producer.SpanContext.TraceID())
	assert.Equal(t, traceID, consumer.SpanContext.TraceID())
	assert.Equal(t, traceID, hsc.TraceID())
	assert.Equal(t, consumer.SpanContext.SpanID(), hsc.SpanID())

	assert.Equal(t, parent.SpanContext().SpanID(), producer.Parent.SpanID())
	assert.Equal(t, producer.SpanContext.SpanID(), consumer.Parent.SpanID())
	assert.True(t, consumer.Parent.IsRemote())

	assert.Equal(t, trace.SpanKindProducer, producer.SpanKind)
	assert.Equal(t, trace.SpanKindConsumer, consumer.SpanKind)

	for _, s := range []tracetest.SpanStub{producer, consumer} {
		assert.Equal(t, "redis", requireAttr(t, s, "messaging.system").AsString())
		assert.Equal(t, f.stream, requireAttr(t, s, "messaging.destination.name").AsString())
		assert.Equal(t, event.EventID(), requireAttr(t, s, "messaging.message.id").AsString())
	}

	assert.Equal(t, "send", requireAttr(t, producer, "messaging.operation.type").AsString())
	assert.Equal(t, "process", requireAttr(t, consumer, "messaging.operation.type").AsString())
	assert.Equal(t, "trace-group", requireAttr(t, consumer, "messaging.consumer.group.name").AsString())
	assert.Equal(t, "promotion.created", requireAttr(t, consumer, "eventbus.event.type").AsString())
	assert.EqualValues(t, 1, requireAttr(t, consumer, "eventbus.attempt").AsInt64())
	assert.Equal(t, codes.Unset, consumer.Status.Code)
}

func TestTracing_PublishBatch(t *testing.T) {
	f := newTracingFixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const n = 3

	events := make([]eventbus.Event, n)
	for i := range events {
		events[i] = testutil.NewTestEvent("promotion.created", map[string]any{"i": i})
	}

	require.NoError(t, f.pub.PublishBatch(ctx, f.stream, events))

	spans := f.exp.GetSpans()
	require.Len(t, spans, n)

	msgs, err := f.client.XRange(ctx, f.stream, "-", "+").Result()
	require.NoError(t, err)
	require.Len(t, msgs, n)

	seen := map[string]bool{}

	for i, msg := range msgs {
		tp, _ := f.metadata(t, msg)["traceparent"].(string)
		require.NotEmpty(t, tp)
		assert.False(t, seen[tp], "traceparent reused")
		seen[tp] = true

		var matched bool

		for _, s := range spans {
			if strings.Contains(tp, s.SpanContext.SpanID().String()) {
				matched = true

				assert.Equal(t, events[i].EventID(), requireAttr(t, s, "messaging.message.id").AsString())
				assert.Equal(t, f.stream+".publish", s.Name)
				assert.Equal(t, trace.SpanKindProducer, s.SpanKind)
			}
		}

		assert.True(t, matched, "no producer span matches %s", tp)
	}
}

func TestTracing_MessageWithoutTraceparentStartsRoot(t *testing.T) {
	f := newTracingFixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	metadataJSON, err := json.Marshal(map[string]any{
		"id": "evt-1", "type": "user.registered", "timestamp": time.Now().UTC().Format(time.RFC3339),
		"version": "1.0", "attempt": 1,
	})
	require.NoError(t, err)
	require.NoError(t, f.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: f.stream, Values: map[string]any{"metadata": string(metadataJSON), "payload": "{}"},
	}).Err())

	// The Subscribe context carries an unrelated span: the message must not join it.
	workerCtx, worker := f.tp.Tracer("test").Start(ctx, "worker")
	defer worker.End()

	stop := f.subscribe(workerCtx, eventbus.SubscriptionConfig{
		Handler: func(context.Context, eventbus.Event) error { return nil },
	})
	defer stop()

	consumer := waitSpans(t, ctx, f.exp, f.stream+".process", 1)[0]

	assert.False(t, consumer.Parent.IsValid())
	assert.NotEqual(t, worker.SpanContext().TraceID(), consumer.SpanContext.TraceID())
}

func TestTracing_HandlerErrorAndRetryKeepProducerTrace(t *testing.T) {
	f := newTracingFixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	require.NoError(t, f.pub.Publish(ctx, f.stream, testutil.NewTestEvent("promotion.created", map[string]any{"k": "v"})))

	var calls int
	var mu sync.Mutex

	stop := f.subscribe(ctx, eventbus.SubscriptionConfig{
		Handler: func(context.Context, eventbus.Event) error {
			mu.Lock()
			defer mu.Unlock()

			calls++
			if calls == 1 {
				return assert.AnError
			}

			return nil
		},
	})
	defer stop()

	producer := waitSpans(t, ctx, f.exp, f.stream+".publish", 1)[0]
	consumers := waitSpans(t, ctx, f.exp, f.stream+".process", 2)

	byAttempt := map[int64]tracetest.SpanStub{}
	for _, s := range consumers {
		byAttempt[requireAttr(t, s, "eventbus.attempt").AsInt64()] = s
	}

	first, second := byAttempt[1], byAttempt[2]
	require.NotZero(t, first.Name)
	require.NotZero(t, second.Name)

	assert.Equal(t, codes.Error, first.Status.Code)
	assert.Equal(t, assert.AnError.Error(), first.Status.Description)
	assert.True(t, hasExceptionEvent(first))
	assert.Equal(t, codes.Unset, second.Status.Code)

	for _, s := range []tracetest.SpanStub{first, second} {
		assert.Equal(t, producer.SpanContext.TraceID(), s.SpanContext.TraceID())
		assert.Equal(t, producer.SpanContext.SpanID(), s.Parent.SpanID(), "attempts are siblings under the producer")
	}

	msgs, err := f.client.XRange(ctx, f.stream, "-", "+").Result()
	require.NoError(t, err)
	require.Len(t, msgs, 2)

	want := fmt.Sprintf("00-%s-%s-01", producer.SpanContext.TraceID(), producer.SpanContext.SpanID())
	assert.Equal(t, want, f.metadata(t, msgs[0])["traceparent"])
	assert.Equal(t, want, f.metadata(t, msgs[1])["traceparent"], "requeued message keeps the producer traceparent")
	assert.EqualValues(t, 2, f.metadata(t, msgs[1])["attempt"])
}

func TestTracing_PublishErrorFailsSpan(t *testing.T) {
	f := newTracingFixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// WRONGTYPE: the stream key holds a string.
	require.NoError(t, f.client.Set(ctx, f.stream, "not-a-stream", 0).Err())

	err := f.pub.Publish(ctx, f.stream, testutil.NewTestEvent("promotion.created", map[string]any{"k": "v"}))
	require.ErrorIs(t, err, eventbus.ErrPublishFailed)

	spans := f.exp.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
	assert.True(t, hasExceptionEvent(spans[0]))

	f.exp.Reset()

	err = f.pub.PublishBatch(ctx, f.stream, []eventbus.Event{
		testutil.NewTestEvent("promotion.created", map[string]any{"k": 1}),
		testutil.NewTestEvent("promotion.created", map[string]any{"k": 2}),
	})
	require.ErrorIs(t, err, eventbus.ErrPublishFailed)

	spans = f.exp.GetSpans()
	require.Len(t, spans, 2)

	for _, s := range spans {
		assert.Equal(t, codes.Error, s.Status.Code)
	}
}
