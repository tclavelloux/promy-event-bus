//nolint:all // Test file
package eventbus_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tclavelloux/promy-event-bus/eventbus"
)

// Compile-time assertion that BaseEvent satisfies the Event interface.
var _ eventbus.Event = eventbus.BaseEvent{}

func TestNewBaseEvent(t *testing.T) {
	event := eventbus.NewBaseEvent("user.registered", "promy-user")

	_, err := uuid.Parse(event.ID)
	assert.NoError(t, err)
	assert.Equal(t, "user.registered", event.Type)
	assert.Equal(t, "promy-user", event.Source)
	assert.Equal(t, "1.0", event.Version)
	assert.Equal(t, time.UTC, event.Timestamp.Location())
	assert.WithinDuration(t, time.Now().UTC(), event.Timestamp, 2*time.Second)
}

func TestBaseEvent_Accessors(t *testing.T) {
	event := eventbus.NewBaseEvent("user.registered", "promy-user")

	assert.Equal(t, event.Type, event.EventType())
	assert.Equal(t, event.ID, event.EventID())
	assert.Equal(t, event.Timestamp, event.EventTime())
	assert.Equal(t, "", event.Data())
}

func TestBaseEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*eventbus.BaseEvent)
		wantErr bool
	}{
		{"valid", func(e *eventbus.BaseEvent) {}, false},
		{"empty ID", func(e *eventbus.BaseEvent) { e.ID = "" }, true},
		{"empty Type", func(e *eventbus.BaseEvent) { e.Type = "" }, true},
		{"zero Timestamp", func(e *eventbus.BaseEvent) { e.Timestamp = time.Time{} }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := eventbus.NewBaseEvent("user.registered", "promy-user")
			tt.modify(&event)

			err := event.Validate()
			if tt.wantErr {
				assert.ErrorIs(t, err, eventbus.ErrInvalidEvent)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
