package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/messaging"
	"github.com/nats-io/nats.go"
)

func Init() messaging.EventBus{
	// Get the database configuration from the config package
	var cfg Config
	err := config.LoadConfig("application", []string{"."}, &cfg)
	if err != nil {
		log.Fatalf("Nats: error loading config %v", err)
	}
	// Initialize NATS connection
	log.Printf("Nats: Connecting to nats on %s \n", cfg.Messaging.Nats.URL)
	eventBus, err := NewEventBus(cfg.Messaging.Nats.URL, cfg.Messaging.Nats.Username, cfg.Messaging.Nats.Password)
	if err != nil {
		log.Fatalf("Nats: error initializing event bus. %v", err)
	}

	return eventBus
}
type NatsEventBus struct {
	nc     *nats.Conn
	mu     sync.RWMutex // Protects the closed flag
	closed bool
}

// NewNatsEventBus creates a new NatsEventBus with automatic reconnection.
func NewEventBus(natsURL, username, password string) (*NatsEventBus, error) {
	nc, err := nats.Connect(natsURL,
		nats.UserInfo(username, password),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			fmt.Printf("Error in subscription: %v\n", err)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server. %v", err)
	}

	return &NatsEventBus{nc: nc}, nil
}

// Publish sends an event to all subscribers of the specified event type.
func (b *NatsEventBus) Publish(ctx context.Context, topic string, event messaging.Event) error {
	if b.isClosed() {
		return errors.New("eventbus is closed")
	}

    // validate topic
    if topic == "" {
        return errors.New("topic must not be empty")
    }

	// Validate the event before publishing
	if event.Type == "" || event.ID == "" {
		return errors.New("event must have a valid ID and Type")
	}

	data, err := encodeEvent(event)
	if err != nil {
		return fmt.Errorf("failed to encode event: %w", err)
	}

	err = b.nc.Publish(topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Subscribe registers a handler for the specified event type and returns an unsubscribe function.
func (b *NatsEventBus) Subscribe(topic string, handler func(ctx context.Context, event messaging.Event), options ...messaging.SubscriptionOption) error {
    if b.isClosed() {
        return errors.New("eventbus is closed")
    }

    // Validate topic before subscribing
    if topic == "" {
        return errors.New("topic must not be empty")
    }
    
    // Process subscription options for consumer group and name.
    opts := messaging.SubscriptionOptions{}
    for _, o := range options {
        o(&opts)
    }
    
    // Log subscriber details if a consumer name is provided.
    if opts.Name != "" {
        log.Printf("Nats: Subscriber '%s'", opts.Name)
    }

    // If a consumer group is specified, use QueueSubscribe to load balance the messages.
    if opts.Group != "" {
        log.Printf("Nats: Joining consumer group '%s' on topic '%s'", opts.Group, topic)
        _, err := b.nc.QueueSubscribe(topic, opts.Group, func(msg *nats.Msg) {
            event, err := decodeEvent(msg.Data)
            if err != nil {
                log.Printf("failed to decode event: %v", err)
                return
            }
            go handler(context.Background(), event)
        })
        if err != nil {
            return fmt.Errorf("failed to subscribe to event: %w", err)
        }
        return nil
    }

    // Otherwise, use normal Subscribe.
    _, err := b.nc.Subscribe(topic, func(msg *nats.Msg) {
        event, err := decodeEvent(msg.Data)
        if err != nil {
            log.Printf("failed to decode event: %v", err)
            return
        }
        go handler(context.Background(), event)
    })
    if err != nil {
        return fmt.Errorf("failed to subscribe to event: %w", err)
    }

    return nil
}

// Close shuts down the event bus and ensures no new events are processed.
func (b *NatsEventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return errors.New("eventbus is already closed")
	}

	b.nc.Close()
	b.closed = true
	return nil
}

func (b *NatsEventBus) isClosed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.closed
}

func encodeEvent(event messaging.Event) ([]byte, error) {
	return json.Marshal(event)
}

func decodeEvent(data []byte) (messaging.Event, error) {
	var evt messaging.Event
	err := json.Unmarshal(data, &evt)
	return evt, err
}
