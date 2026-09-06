//nolint:all // Test file
package eventbus_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tclavelloux/promy-event-bus/eventbus"
)

func TestGetValidator_Singleton(t *testing.T) {
	first := eventbus.GetValidator()
	second := eventbus.GetValidator()

	assert.Same(t, first, second)
}

func TestValidateStruct_Valid(t *testing.T) {
	event := eventbus.NewBaseEvent("user.registered", "promy-user")

	assert.NoError(t, eventbus.ValidateStruct(event))
}

func TestValidateStruct_NonStruct(t *testing.T) {
	err := eventbus.ValidateStruct(42)

	assert.ErrorIs(t, err, eventbus.ErrInvalidEvent)
	assert.NotContains(t, err.Error(), "field '")
}

func TestValidateStruct_MessageByTag(t *testing.T) {
	type emailStruct struct {
		Email string `validate:"email"`
	}
	type minStruct struct {
		N int `validate:"min=5"`
	}
	type maxStruct struct {
		N int `validate:"max=5"`
	}
	type defaultStruct struct {
		S string `validate:"alphanum"`
	}

	validEvent := eventbus.NewBaseEvent("user.registered", "promy-user")

	uuidEvent := validEvent
	uuidEvent.ID = "not-a-uuid"

	tests := []struct {
		name    string
		input   any
		wantMsg string
	}{
		{"required", eventbus.BaseEvent{}, "field 'ID' is required"},
		{"uuid", uuidEvent, "field 'ID' must be a valid UUID"},
		{"email", emailStruct{Email: "nope"}, "field 'Email' must be a valid email"},
		{"min", minStruct{N: 1}, "field 'N' must be at least 5"},
		{"max", maxStruct{N: 9}, "field 'N' must be at most 5"},
		{"default", defaultStruct{S: "!!"}, "field 'S' failed validation 'alphanum'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eventbus.ValidateStruct(tt.input)

			assert.ErrorIs(t, err, eventbus.ErrInvalidEvent)
			assert.Contains(t, err.Error(), tt.wantMsg)
		})
	}
}
