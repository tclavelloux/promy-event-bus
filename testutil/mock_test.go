package testutil_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/testutil"
)

var (
	_ eventbus.EventPublisher  = (*testutil.MockPublisher)(nil)
	_ eventbus.EventSubscriber = (*testutil.MockSubscriber)(nil)
)

func TestMockPublisher(t *testing.T) {
	ctx := context.Background()
	event := testutil.NewTestEvent("user.registered", map[string]any{"user_id": "u-1"})

	tests := []struct {
		name     string
		setup    func(*testutil.MockPublisher)
		call     func(*testutil.MockPublisher) error
		expected string
	}{
		{
			name: "Publish",
			setup: func(m *testutil.MockPublisher) {
				m.On("Publish", ctx, "events:users", event).Return(assert.AnError)
			},
			call: func(m *testutil.MockPublisher) error {
				return m.Publish(ctx, "events:users", event)
			},
		},
		{
			name: "PublishBatch",
			setup: func(m *testutil.MockPublisher) {
				m.On("PublishBatch", ctx, "events:users", []eventbus.Event{event}).Return(assert.AnError)
			},
			call: func(m *testutil.MockPublisher) error {
				return m.PublishBatch(ctx, "events:users", []eventbus.Event{event})
			},
		},
		{
			name: "Close",
			setup: func(m *testutil.MockPublisher) {
				m.On("Close").Return(assert.AnError)
			},
			call: func(m *testutil.MockPublisher) error {
				return m.Close()
			},
		},
		{
			name: "Health",
			setup: func(m *testutil.MockPublisher) {
				m.On("Health", ctx).Return(assert.AnError)
			},
			call: func(m *testutil.MockPublisher) error {
				return m.Health(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(testutil.MockPublisher)
			tt.setup(m)

			err := tt.call(m)

			assert.ErrorIs(t, err, assert.AnError)
			m.AssertExpectations(t)
		})
	}
}

func TestMockSubscriber(t *testing.T) {
	ctx := context.Background()
	subConfig := eventbus.SubscriptionConfig{Stream: "events:users", ConsumerGroup: "g", ConsumerID: "c"}

	tests := []struct {
		name  string
		setup func(*testutil.MockSubscriber)
		call  func(*testutil.MockSubscriber) error
	}{
		{
			name: "Subscribe",
			setup: func(m *testutil.MockSubscriber) {
				m.On("Subscribe", ctx, subConfig).Return(assert.AnError)
			},
			call: func(m *testutil.MockSubscriber) error {
				return m.Subscribe(ctx, subConfig)
			},
		},
		{
			name: "Close",
			setup: func(m *testutil.MockSubscriber) {
				m.On("Close").Return(assert.AnError)
			},
			call: func(m *testutil.MockSubscriber) error {
				return m.Close()
			},
		},
		{
			name: "Health",
			setup: func(m *testutil.MockSubscriber) {
				m.On("Health", ctx).Return(assert.AnError)
			},
			call: func(m *testutil.MockSubscriber) error {
				return m.Health(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(testutil.MockSubscriber)
			tt.setup(m)

			err := tt.call(m)

			assert.ErrorIs(t, err, assert.AnError)
			m.AssertExpectations(t)
		})
	}
}
