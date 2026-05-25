package eventbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DLQEntry represents a failed event routed to the Dead Letter Queue.
// It implements the Event interface so it can be published to StreamDLQ
// using the standard Publisher.
type DLQEntry struct {
	id string

	OriginalStream    string    `json:"original_stream"    validate:"required"`
	OriginalEventID   string    `json:"original_event_id"  validate:"required"`
	OriginalEventType string    `json:"original_event_type" validate:"required"`
	OriginalPayload   string    `json:"original_payload"`
	FailureReason     string    `json:"failure_reason"     validate:"required"`
	FailedAt          time.Time `json:"failed_at"          validate:"required"`
	FailedService     string    `json:"failed_service"     validate:"required"`
	AttemptsExhausted int       `json:"attempts_exhausted" validate:"required,min=1"`
}

// NewDLQEntry creates a DLQ entry from a failed event.
func NewDLQEntry(stream string, event Event, err error, service string, attempts int) *DLQEntry {
	now := time.Now().UTC()

	return &DLQEntry{
		id:                uuid.New().String(),
		OriginalStream:    stream,
		OriginalEventID:   event.EventID(),
		OriginalEventType: event.EventType(),
		OriginalPayload:   event.Data(),
		FailureReason:     err.Error(),
		FailedAt:          now,
		FailedService:     service,
		AttemptsExhausted: attempts,
	}
}

func (d *DLQEntry) EventID() string      { return d.id }
func (d *DLQEntry) EventType() string    { return "dlq." + d.OriginalEventType }
func (d *DLQEntry) EventTime() time.Time { return d.FailedAt }

func (d *DLQEntry) Data() string {
	b, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}

	return string(b)
}

func (d *DLQEntry) Validate() error {
	if d.OriginalStream == "" {
		return ErrInvalidEvent
	}
	if d.OriginalEventID == "" {
		return ErrInvalidEvent
	}
	if d.OriginalEventType == "" {
		return ErrInvalidEvent
	}
	if d.FailureReason == "" {
		return ErrInvalidEvent
	}
	if d.FailedService == "" {
		return ErrInvalidEvent
	}
	if d.AttemptsExhausted < 1 {
		return ErrInvalidEvent
	}

	return nil
}
