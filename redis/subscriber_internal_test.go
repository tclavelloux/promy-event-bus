package redis

import (
	"context"
	"encoding/json"
	"testing"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequeue_MetadataMarshalFailureIsReported(t *testing.T) {
	var got []eventbus.SubscriberError

	config := eventbus.SubscriptionConfig{
		Stream:        "events:test",
		ConsumerGroup: "group",
		ConsumerID:    "consumer",
		OnError: func(_ context.Context, e eventbus.SubscriberError) {
			got = append(got, e)
		},
	}

	// Nil client: marshal fails first, so Redis is never touched.
	(&Subscriber{}).requeue(
		context.Background(), config, goredis.XMessage{ID: "1-0"},
		map[string]any{"bad": make(chan int)}, 2, "evt-1", "user.registered",
	)

	require.Len(t, got, 1)
	assert.Equal(t, eventbus.StageMetadataMarshal, got[0].Stage)
	assert.Equal(t, "events:test", got[0].Stream)
	assert.Equal(t, "group", got[0].ConsumerGroup)
	assert.Equal(t, "consumer", got[0].ConsumerID)
	assert.Equal(t, "1-0", got[0].MessageID)
	assert.Equal(t, "evt-1", got[0].EventID)
	assert.Equal(t, "user.registered", got[0].EventType)
	assert.Equal(t, 2, got[0].Attempt)

	var unsupported *json.UnsupportedTypeError
	assert.ErrorAs(t, got[0], &unsupported)
}
