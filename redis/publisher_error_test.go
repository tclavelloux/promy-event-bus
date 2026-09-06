//nolint:all // Test file
package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
	"github.com/tclavelloux/promy-event-bus/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failingValidateEvent has no struct-tag validation constraints so it always
// clears ValidateStruct, but its business-rule Validate() always fails.
type failingValidateEvent struct {
	eventbus.BaseEvent
}

func newFailingValidateEvent() *failingValidateEvent {
	return &failingValidateEvent{BaseEvent: eventbus.NewBaseEvent("test.failing", "test")}
}

func (e *failingValidateEvent) Validate() error {
	return errValidateSentinel
}

var errValidateSentinel = errors.New("business validation always fails")

// unmarshalableEvent passes both validation layers but cannot be marshaled to
// JSON as its payload, since MarshalJSON always errors.
type unmarshalableEvent struct {
	eventbus.BaseEvent
}

func newUnmarshalableEvent() *unmarshalableEvent {
	return &unmarshalableEvent{BaseEvent: eventbus.NewBaseEvent("test.unmarshalable", "test")}
}

func (e *unmarshalableEvent) Validate() error {
	return nil
}

func (e *unmarshalableEvent) MarshalJSON() ([]byte, error) {
	return nil, errMarshalSentinel
}

var errMarshalSentinel = errors.New("marshal always fails")

func TestNewPublisher_AppliesAllOptions(t *testing.T) {
	config := eventbus.RedisConfig{
		DSN:             testDSN(t),
		PoolSize:        7,
		MaxRetries:      4,
		MinRetryBackoff: 5 * time.Millisecond,
		MaxRetryBackoff: 200 * time.Millisecond,
		DialTimeout:     2 * time.Second,
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
	}

	publisher, err := redis.NewPublisher(config)
	require.NoError(t, err)
	defer publisher.Close()

	assert.NoError(t, publisher.Health(t.Context()))
}

func TestPublisher_Publish_StructTagValidationFails(t *testing.T) {
	config := eventbus.RedisConfig{DSN: testDSN(t)}
	publisher, err := redis.NewPublisher(config)
	require.NoError(t, err)
	defer publisher.Close()

	err = publisher.Publish(context.Background(), uniqueStream(t, "pub"), eventbus.BaseEvent{})

	assert.ErrorIs(t, err, eventbus.ErrInvalidEvent)
}

func TestPublisher_Publish_BusinessValidationFails(t *testing.T) {
	config := eventbus.RedisConfig{DSN: testDSN(t)}
	publisher, err := redis.NewPublisher(config)
	require.NoError(t, err)
	defer publisher.Close()

	err = publisher.Publish(context.Background(), uniqueStream(t, "pub"), newFailingValidateEvent())

	assert.ErrorIs(t, err, errValidateSentinel)
}

func TestPublisher_Publish_PayloadMarshalFails(t *testing.T) {
	config := eventbus.RedisConfig{DSN: testDSN(t)}
	publisher, err := redis.NewPublisher(config)
	require.NoError(t, err)
	defer publisher.Close()

	err = publisher.Publish(context.Background(), uniqueStream(t, "pub"), newUnmarshalableEvent())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal payload")
}

func TestPublisher_Publish_XAddFails(t *testing.T) {
	config := eventbus.RedisConfig{DSN: testDSN(t)}
	publisher, err := redis.NewPublisher(config)
	require.NoError(t, err)
	require.NoError(t, publisher.Close())

	event := testutil.NewTestEvent("user.registered", map[string]any{"user_id": "u-1"})
	err = publisher.Publish(context.Background(), uniqueStream(t, "pub"), event)

	assert.ErrorIs(t, err, eventbus.ErrPublishFailed)
}

func TestPublisher_PublishBatch_Errors(t *testing.T) {
	config := eventbus.RedisConfig{DSN: testDSN(t)}

	tests := []struct {
		name       string
		batch      func(t *testing.T) []eventbus.Event
		wantErr    error
		wantErrMsg string
		closeFirst bool
	}{
		{
			name: "struct tag validation fails",
			batch: func(t *testing.T) []eventbus.Event {
				return []eventbus.Event{eventbus.BaseEvent{}}
			},
			wantErr: eventbus.ErrInvalidEvent,
		},
		{
			name: "business validation fails",
			batch: func(t *testing.T) []eventbus.Event {
				return []eventbus.Event{newFailingValidateEvent()}
			},
			wantErr: errValidateSentinel,
		},
		{
			name: "payload marshal fails",
			batch: func(t *testing.T) []eventbus.Event {
				return []eventbus.Event{newUnmarshalableEvent()}
			},
			wantErrMsg: "failed to marshal payload",
		},
		{
			name: "pipe exec fails on closed client",
			batch: func(t *testing.T) []eventbus.Event {
				return []eventbus.Event{testutil.NewTestEvent("user.registered", map[string]any{"user_id": "u-1"})}
			},
			wantErr:    eventbus.ErrPublishFailed,
			closeFirst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher, err := redis.NewPublisher(config)
			require.NoError(t, err)

			if tt.closeFirst {
				require.NoError(t, publisher.Close())
			} else {
				defer publisher.Close()
			}

			err = publisher.PublishBatch(context.Background(), uniqueStream(t, "pub-batch"), tt.batch(t))

			require.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			}
		})
	}
}
