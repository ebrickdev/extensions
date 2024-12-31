package inmemory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ebrickdev/ebrick/event"
	"github.com/stretchr/testify/assert"
)

// waitForHandler waits for the WaitGroup to be done or times out.
func waitForHandler(t *testing.T, wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for event handler")
	}
}

func TestSubscribeAndPublish(t *testing.T) {
	bus, _ := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// Subscribe to "test.event"
	err := bus.Subscribe("test.event", func(ctx context.Context, e event.Event) {
		assert.Equal(t, "test.event", e.Type, "Event type mismatch")
		assert.Equal(t, "source-1", e.Source, "Event source mismatch")
		assert.Equal(t, map[string]any{"key": "value"}, e.Data, "Event data mismatch")
		wg.Done()
	})
	assert.NoError(t, err, "Subscription should not return an error")

	// Publish the event
	ctx := context.Background()
	event := event.Event{
		ID:          "1",
		Source:      "source-1",
		SpecVersion: "1.0",
		Type:        "test.event",
		Data:        map[string]any{"key": "value"},
		Time:        time.Now(),
	}
	err = bus.Publish(ctx, event)
	assert.NoError(t, err, "Publish should not return an error")

	// Wait for the handler to process the event
	waitForHandler(t, &wg)
}

func TestMultipleSubscribers(t *testing.T) {
	bus, _ := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Subscriber 1
	err := bus.Subscribe("multi.event", func(ctx context.Context, e event.Event) {
		assert.Equal(t, "multi.event", e.Type, "Event type mismatch for subscriber 1")
		assert.Equal(t, map[string]any{"user": "john"}, e.Data, "Event data mismatch for subscriber 1")
		wg.Done()
	})
	assert.NoError(t, err, "Subscription 1 should not return an error")

	// Subscriber 2
	err = bus.Subscribe("multi.event", func(ctx context.Context, e event.Event) {
		assert.Equal(t, "multi.event", e.Type, "Event type mismatch for subscriber 2")
		assert.Equal(t, map[string]any{"user": "john"}, e.Data, "Event data mismatch for subscriber 2")
		wg.Done()
	})
	assert.NoError(t, err, "Subscription 2 should not return an error")

	// Publish the event
	ctx := context.Background()
	event := event.Event{
		ID:          "2",
		Source:      "source-2",
		SpecVersion: "1.0",
		Type:        "multi.event",
		Data:        map[string]any{"user": "john"},
		Time:        time.Now(),
	}
	err = bus.Publish(ctx, event)
	assert.NoError(t, err, "Publish should not return an error")

	// Wait for both handlers to process the event
	waitForHandler(t, &wg)
}

func TestClose(t *testing.T) {
	bus, _ := NewEventBus()

	// Subscribe to an event
	err := bus.Subscribe("test.event", func(ctx context.Context, e event.Event) {
		assert.Fail(t, "Handler should not be called after bus is closed")
	})
	assert.NoError(t, err, "Subscription should not return an error")

	// Close the bus
	err = bus.Close()
	assert.NoError(t, err, "Close should not return an error")

	// Attempt to publish after closing
	err = bus.Publish(context.Background(), event.Event{
		ID:          "3",
		Source:      "source-3",
		SpecVersion: "1.0",
		Type:        "test.event",
		Data:        map[string]any{"key": "value"},
		Time:        time.Now(),
	})
	assert.Error(t, err, "Publish should return an error after bus is closed")
	assert.Equal(t, "eventbus is closed", err.Error(), "Error message mismatch")
}

func TestSubscribeAfterClose(t *testing.T) {
	bus, _ := NewEventBus()

	// Close the bus
	err := bus.Close()
	assert.NoError(t, err, "Close should not return an error")

	// Attempt to subscribe after closing
	err = bus.Subscribe("test.event", func(ctx context.Context, e event.Event) {})
	assert.Error(t, err, "Subscribe should return an error after bus is closed")
	assert.Equal(t, "eventbus is closed", err.Error(), "Error message mismatch")
}

func TestContextCancellation(t *testing.T) {
	bus, _ := NewEventBus()
	defer bus.Close()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	// Subscribe to "cancel.event"
	err := bus.Subscribe("cancel.event", func(ctx context.Context, e event.Event) {
		select {
		case <-ctx.Done():
			// Ensure the handler respects context cancellation
			wg.Done()
		default:
			assert.Fail(t, "Handler did not respect context cancellation")
		}
	})
	assert.NoError(t, err, "Subscription should not return an error")

	// Cancel the context before publishing
	cancel()

	// Publish an event with a canceled context
	err = bus.Publish(ctx, event.Event{
		ID:          "4",
		Source:      "source-4",
		SpecVersion: "1.0",
		Type:        "cancel.event",
		Data:        map[string]any{"key": "value"},
		Time:        time.Now(),
	})
	assert.NoError(t, err, "Publish should not return an error")

	// Wait for the handler to process the cancellation
	waitForHandler(t, &wg)
}

func TestDuplicateClose(t *testing.T) {
	bus, _ := NewEventBus()

	// Close the bus once
	err := bus.Close()
	assert.NoError(t, err, "First close should not return an error")

	// Attempt to close it again
	err = bus.Close()
	assert.Error(t, err, "Second close should return an error")
	assert.Equal(t, "eventbus is already closed", err.Error(), "Error message mismatch")
}
