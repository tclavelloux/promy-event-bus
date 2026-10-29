package eventbus_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tclavelloux/promy-event-bus/eventbus"
)

func TestSubscriberError_ErrorAndUnwrap(t *testing.T) {
	cause := errors.New("boom")
	err := eventbus.SubscriberError{
		Stage:     eventbus.StageAck,
		Stream:    "events:users",
		MessageID: "1-0",
		Err:       cause,
	}

	assert.Equal(t, "eventbus subscriber: ack failed for message 1-0 on events:users: boom", err.Error())
	assert.Equal(t, cause, err.Unwrap())
	assert.ErrorIs(t, err, cause)
}
