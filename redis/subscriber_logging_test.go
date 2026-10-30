package redis_test

import (
	"context"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newLoggingFixture builds an onErrorFixture whose subscriber logs to buf and traces to a fresh provider.
func newLoggingFixture(t *testing.T, attempt int, buf *lockedBuffer) *onErrorFixture {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := newOnErrorFixture(t, attempt)
	_ = f.sub.Close()

	_, tp := newTracing(t)
	logger := zerolog.New(buf)

	sub, err := redis.NewSubscriber(
		eventbus.Config{Redis: eventbus.RedisConfig{DSN: testDSN(t)}},
		redis.WithLogger(&logger), redis.WithTracerProvider(tp),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sub.Close() })

	f.sub = sub

	return f
}

func TestSubscriber_Logging_MalformedMetadataWarns(t *testing.T) {
	buf := &lockedBuffer{}
	f := newLoggingFixture(t, 1, buf)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	badID, err := f.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: f.stream, Values: map[string]any{"metadata": "{not json", "payload": "{}"},
	}).Result()
	require.NoError(t, err)

	f.subscribe(ctx, eventbus.SubscriptionConfig{
		Handler: func(context.Context, eventbus.Event) error { return nil },
	})

	var line map[string]any

	for _, l := range buf.waitLines(t, ctx, 1) {
		if l["messageId"] == badID {
			line = l
		}
	}

	require.NotNil(t, line)
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "event bus subscriber swallowed error", line["message"])
	assert.Equal(t, "metadata_unmarshal", line["stage"])
	assert.Equal(t, f.stream, line["stream"])
	assert.Equal(t, f.group, line["consumerGroup"])
	assert.Equal(t, f.consumer, line["consumerId"])
	assert.NotEmpty(t, line["error"])
	assert.NotContains(t, line, "traceId", "no span exists before metadata decodes")
}

func TestSubscriber_Logging_ErrorStageLogsWithoutOnError(t *testing.T) {
	buf := &lockedBuffer{}
	f := newLoggingFixture(t, 3, buf)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	f.subscribe(ctx, eventbus.SubscriptionConfig{
		DLQPublisher: failingPublisher{},
		Handler:      func(context.Context, eventbus.Event) error { return assert.AnError },
	})

	lines := buf.waitLines(t, ctx, 1)
	require.Len(t, lines, 1)

	line := lines[0]
	assert.Equal(t, "error", line["level"])
	assert.Equal(t, "event bus subscriber swallowed error", line["message"])
	assert.Equal(t, "dlq_publish", line["stage"])
	assert.Equal(t, f.eventID, line["eventId"])
	assert.Equal(t, "user.registered", line["eventType"])
	assert.EqualValues(t, 3, line["attempt"])
	assert.Equal(t, f.messageID, line["messageId"])
	assert.Equal(t, errDLQDown.Error(), line["error"])
	assert.Len(t, line["traceId"], 32)
	assert.Len(t, line["spanId"], 16)
}

func TestSubscriber_Logging_OnErrorReplacesLogger(t *testing.T) {
	buf := &lockedBuffer{}
	f := newLoggingFixture(t, 3, buf)
	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	f.subscribe(ctx, eventbus.SubscriptionConfig{
		DLQPublisher: failingPublisher{},
		OnError:      col.record,
		Handler:      func(context.Context, eventbus.Event) error { return assert.AnError },
	})

	col.waitFor(t, ctx, 1)
	<-ctx.Done()

	assert.Empty(t, buf.lines(t), "logger must not be used for errors reported through OnError")
}

func TestSubscriber_Logging_DropWithoutDLQWarns(t *testing.T) {
	buf := &lockedBuffer{}
	f := newLoggingFixture(t, 3, buf)
	col := &errCollector{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	f.subscribe(ctx, eventbus.SubscriptionConfig{
		OnError: col.record, // the drop is not an OnError stage: it is logged regardless
		Handler: func(context.Context, eventbus.Event) error { return assert.AnError },
	})

	lines := buf.waitLines(t, ctx, 1)
	require.Len(t, lines, 1)

	line := lines[0]
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "event dropped after max attempts without dlq", line["message"])
	assert.Equal(t, f.eventID, line["eventId"])
	assert.Equal(t, "user.registered", line["eventType"])
	assert.EqualValues(t, 3, line["attempt"])
	assert.Equal(t, f.stream, line["stream"])
	assert.Equal(t, f.messageID, line["messageId"])
	assert.NotContains(t, line, "stage")
	assert.NotContains(t, line, "error")
	assert.Len(t, line["traceId"], 32)
	assert.Empty(t, col.snapshot())
}
