package testutil

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TestEvent is a concrete Event implementation for testing purposes.
// Downstream services define their own event structs; this is only for
// promy-event-bus internal tests and examples.
//
// MarshalJSON produces a flat JSON structure where Payload fields are merged
// at the top level alongside metadata — matching how real service event structs
// serialize (embedded BaseEvent + domain fields).
type TestEvent struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	CreatedAt time.Time      `json:"timestamp"`
	Version   string         `json:"version"`
	Source    string         `json:"source"`
	Payload   map[string]any `json:"-"`
}

// NewTestEvent creates a test event with the given type and payload fields.
func NewTestEvent(eventType string, payload map[string]any) *TestEvent {
	return &TestEvent{
		ID:        uuid.NewString(),
		Type:      eventType,
		CreatedAt: time.Now().UTC(),
		Version:   "1.0",
		Source:    "test",
		Payload:   payload,
	}
}

func (e *TestEvent) EventType() string    { return e.Type }
func (e *TestEvent) EventID() string      { return e.ID }
func (e *TestEvent) EventTime() time.Time { return e.CreatedAt }
func (e *TestEvent) Validate() error      { return nil }

func (e *TestEvent) Data() string {
	b, _ := json.Marshal(e.Payload) //nolint:errchkjson // test helper

	return string(b)
}

// MarshalJSON flattens metadata + payload into a single JSON object.
func (e *TestEvent) MarshalJSON() ([]byte, error) {
	flat := make(map[string]any, len(e.Payload)+5)
	flat["id"] = e.ID
	flat["type"] = e.Type
	flat["timestamp"] = e.CreatedAt.Format(time.RFC3339Nano)
	flat["version"] = e.Version
	flat["source"] = e.Source
	for k, v := range e.Payload {
		flat[k] = v
	}

	return json.Marshal(flat)
}
