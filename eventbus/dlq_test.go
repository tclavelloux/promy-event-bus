package eventbus_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/testutil"
)

const (
	testUserID    = "u-1"
	testUserIDKey = "user_id"
)

func TestNewDLQEntry_Fields(t *testing.T) {
	event := testutil.NewTestEvent("user.registered", map[string]any{testUserIDKey: testUserID})
	handlerErr := errors.New("timeout calling email service")

	entry := eventbus.NewDLQEntry("events:users", event, handlerErr, "promy-crm", 3)

	assert.Equal(t, "events:users", entry.OriginalStream)
	assert.Equal(t, event.EventID(), entry.OriginalEventID)
	assert.Equal(t, "user.registered", entry.OriginalEventType)
	assert.Equal(t, event.Data(), entry.OriginalPayload)
	assert.Equal(t, "timeout calling email service", entry.FailureReason)
	assert.Equal(t, "promy-crm", entry.FailedService)
	assert.Equal(t, 3, entry.AttemptsExhausted)
	assert.WithinDuration(t, time.Now().UTC(), entry.FailedAt, 2*time.Second)
}

func TestDLQEntry_EventInterface(t *testing.T) {
	event := testutil.NewTestEvent("subscription.started", map[string]any{"sub_id": "s-1"})
	entry := eventbus.NewDLQEntry("events:subscriptions", event, errors.New("db error"), "promy-crm", 3)

	t.Run("EventType has dlq prefix", func(t *testing.T) {
		assert.Equal(t, "dlq.subscription.started", entry.EventType())
	})

	t.Run("EventID is a non-empty UUID different from original", func(t *testing.T) {
		assert.NotEmpty(t, entry.EventID())
		assert.NotEqual(t, event.EventID(), entry.EventID())
		assert.Len(t, entry.EventID(), 36) // UUID format
	})

	t.Run("EventTime equals FailedAt", func(t *testing.T) {
		assert.Equal(t, entry.FailedAt, entry.EventTime())
	})
}

func TestDLQEntry_Data_RoundTrip(t *testing.T) {
	event := testutil.NewTestEvent("user.registered", map[string]any{"email": "test@example.com"})
	entry := eventbus.NewDLQEntry("events:users", event, errors.New("connection refused"), "promy-crm", 2)

	data := entry.Data()
	require.NotEmpty(t, data)

	var decoded map[string]any
	err := json.Unmarshal([]byte(data), &decoded)
	require.NoError(t, err)

	assert.Equal(t, "events:users", decoded["original_stream"])
	assert.Equal(t, event.EventID(), decoded["original_event_id"])
	assert.Equal(t, "user.registered", decoded["original_event_type"])
	assert.Equal(t, event.Data(), decoded["original_payload"])
	assert.Equal(t, "connection refused", decoded["failure_reason"])
	assert.Equal(t, "promy-crm", decoded["failed_service"])
	assert.Equal(t, float64(2), decoded["attempts_exhausted"])
	assert.NotEmpty(t, decoded["failed_at"])
}

func TestDLQEntry_Validate_Valid(t *testing.T) {
	event := testutil.NewTestEvent("user.registered", map[string]any{testUserIDKey: testUserID})
	entry := eventbus.NewDLQEntry("events:users", event, errors.New("fail"), "promy-crm", 3)

	assert.NoError(t, entry.Validate())
}

func TestDLQEntry_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*eventbus.DLQEntry)
	}{
		{"empty OriginalStream", func(e *eventbus.DLQEntry) { e.OriginalStream = "" }},
		{"empty OriginalEventID", func(e *eventbus.DLQEntry) { e.OriginalEventID = "" }},
		{"empty OriginalEventType", func(e *eventbus.DLQEntry) { e.OriginalEventType = "" }},
		{"empty FailureReason", func(e *eventbus.DLQEntry) { e.FailureReason = "" }},
		{"empty FailedService", func(e *eventbus.DLQEntry) { e.FailedService = "" }},
		{"zero AttemptsExhausted", func(e *eventbus.DLQEntry) { e.AttemptsExhausted = 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := testutil.NewTestEvent("user.registered", map[string]any{testUserIDKey: testUserID})
			entry := eventbus.NewDLQEntry("events:users", event, errors.New("fail"), "promy-crm", 3)
			tt.modify(entry)

			assert.ErrorIs(t, entry.Validate(), eventbus.ErrInvalidEvent)
		})
	}
}

func TestNewDLQEntry_EmptyData(t *testing.T) {
	base := eventbus.NewBaseEvent("some.type", "test-source")
	entry := eventbus.NewDLQEntry("events:test", &base, errors.New("fail"), "test-svc", 1)

	assert.Equal(t, "", entry.OriginalPayload)
	assert.NoError(t, entry.Validate())
}
