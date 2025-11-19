package testutil

import (
	"context"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"

	"github.com/stretchr/testify/mock"
)

// MockSubscriber is a mock implementation of EventSubscriber for testing.
type MockSubscriber struct {
	mock.Mock
}

// Subscribe mocks the Subscribe method.
func (m *MockSubscriber) Subscribe(ctx context.Context, config eventbus.SubscriptionConfig) error {
	args := m.Called(ctx, config)

	return args.Error(0)
}

// Close mocks the Close method.
func (m *MockSubscriber) Close() error {
	args := m.Called()

	return args.Error(0)
}

// Health mocks the Health method.
func (m *MockSubscriber) Health(ctx context.Context) error {
	args := m.Called(ctx)

	return args.Error(0)
}
