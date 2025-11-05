package testutil

import (
	"context"

	eventbus "github.com/tclavelloux/promy-event-bus"

	"github.com/stretchr/testify/mock"
)

// MockPublisher is a mock implementation of EventPublisher for testing.
type MockPublisher struct {
	mock.Mock
}

// Publish mocks the Publish method.
func (m *MockPublisher) Publish(ctx context.Context, stream string, event eventbus.Event) error {
	args := m.Called(ctx, stream, event)
	return args.Error(0)
}

// PublishBatch mocks the PublishBatch method.
func (m *MockPublisher) PublishBatch(ctx context.Context, stream string, events []eventbus.Event) error {
	args := m.Called(ctx, stream, events)
	return args.Error(0)
}

// Close mocks the Close method.
func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Health mocks the Health method.
func (m *MockPublisher) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
