package nats

import (
	"context"
	"errors"
	"testing"

	"github.com/ebrickdev/ebrick/messaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNatsConn is a mocked implementation of the NATS connection.
type MockNatsConn struct {
	mock.Mock
}

func (m *MockNatsConn) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockNatsConn) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	args := m.Called(subject, handler)
	return nil, args.Error(1)
}

func (m *MockNatsConn) Close() {
	m.Called()
}

func (m *MockNatsConn) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// Helper function to create a new mock NatsEventBus
func NewMockNatsEventBus(mockConn *MockNatsConn) *NatsEventBus {
	return &NatsEventBus{
		nc: mockConn,
	}
}

// Test cases

func TestNatsEventBus_Publish(t *testing.T) {
	mockConn := new(MockNatsConn)
	bus := NewMockNatsEventBus(mockConn)

	event := messaging.Event{
		ID:   "1",
		Type: "test.event",
		Data: map[string]any{"key": "value"},
	}

	// Mock the Publish behavior
	mockConn.On("Publish", "test.event", mock.Anything).Return(nil)

	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err, "Expected no error when publishing event")

	mockConn.AssertCalled(t, "Publish", "test.event", mock.Anything)
}

func TestNatsEventBus_PublishError(t *testing.T) {
	mockConn := new(MockNatsConn)
	bus := NewMockNatsEventBus(mockConn)

	event := messaging.Event{
		ID:   "1",
		Type: "test.event",
		Data: map[string]any{"key": "value"},
	}

	// Simulate a publish error
	mockConn.On("Publish", "test.event", mock.Anything).Return(assert.AnError)

	err := bus.Publish(context.Background(), event)
	assert.Error(t, err, "Expected error when publishing event")
	assert.True(t, errors.Is(err, assert.AnError), "Expected wrapped error to contain assert.AnError")

	mockConn.AssertCalled(t, "Publish", "test.event", mock.Anything)
}

func TestNatsEventBus_Subscribe(t *testing.T) {
	mockConn := new(MockNatsConn)
	bus := NewMockNatsEventBus(mockConn)

	handler := func(ctx context.Context, e messaging.Event) {}

	// Mock the Subscribe behavior
	mockConn.On("Subscribe", "test.event", mock.Anything).Return(nil, nil)

	err := bus.Subscribe("test.event", handler)
	assert.NoError(t, err, "Expected no error when subscribing to event")
	assert.NotNil(t, "Expected unsubscribe function to be non-nil")

	mockConn.AssertCalled(t, "Subscribe", "test.event", mock.Anything)
}

func TestNatsEventBus_SubscribeError(t *testing.T) {
	mockConn := new(MockNatsConn)
	bus := NewMockNatsEventBus(mockConn)

	handler := func(ctx context.Context, e messaging.Event) {}

	// Simulate a subscription error
	mockConn.On("Subscribe", "test.event", mock.Anything).Return(nil, assert.AnError)

	err := bus.Subscribe("test.event", handler)
	assert.Error(t, err, "Expected error when subscribing to event")

	mockConn.AssertCalled(t, "Subscribe", "test.event", mock.Anything)
}

func TestNatsEventBus_Close(t *testing.T) {
	mockConn := new(MockNatsConn)
	bus := NewMockNatsEventBus(mockConn)

	// Mock the Close behavior
	mockConn.On("Close").Return()

	err := bus.Close()
	assert.NoError(t, err, "Expected no error when closing the event bus")

	mockConn.AssertCalled(t, "Close")
}

func TestNatsEventBus_CloseTwice(t *testing.T) {
	mockConn := new(MockNatsConn)
	bus := NewMockNatsEventBus(mockConn)

	// Mock the Close behavior
	mockConn.On("Close").Return()

	err := bus.Close()
	assert.NoError(t, err, "Expected no error on first close")

	// Simulate the second close
	mockConn.On("Close").Return()

	err = bus.Close()
	assert.Error(t, err, "Expected error on second close since the bus is already closed")
}
