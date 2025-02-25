package redisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/messaging"
	"github.com/redis/go-redis/v9"
)

const errorSleepDuration = time.Second

// Init loads configuration and sets up the default event bus.
func Init() *RedisStream {
	var cfg RedisStreamConfig
	err := config.LoadConfigByKey("application", "messaging.redis", []string{"."}, &cfg, map[string]any{})
	if err != nil {
		log.Fatalf("Redis Stream: unable to load Redis config: %v", err)
	}
	return NewRedisStream(&cfg)
}

// RedisStream wraps a Redis client.
type RedisStream struct {
	client *redis.Client
}

// NewRedisStream creates a new RedisStream and verifies the connection.
func NewRedisStream(cfg *RedisStreamConfig) *RedisStream {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.URL,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis Stream: unable to connect to Redis: %v", err)
	}
	log.Println("Redis Stream: Redis Stream initialized successfully")

	return &RedisStream{
		client: client,
	}
}

// Close terminates the Redis connection.
func (r *RedisStream) Close() error {
	log.Println("Closing Redis connection")
	return r.client.Close()
}

// Publish serializes the event as JSON and adds it to the specified stream.
func (r *RedisStream) Publish(ctx context.Context, topic string, event messaging.Event) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Redis Stream: failed to serialize event: %v", err)
		return err
	}

	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"event": string(eventData),
		},
	}).Result()
	if err != nil {
		log.Printf("Redis Stream: failed to publish event: %v", err)
		return err
	}
	return nil
}

// Subscribe subscribes to a Redis stream.
// - If a consumer group is provided (SubscriptionOptions.Group is non-empty),
//   it uses consumer group semantics (with XREADGroup and offset ">").
// - Otherwise, it uses a plain XREAD subscription with the fixed offset "$" for new messages.
// The method accepts a context for cancellation.
func (r *RedisStream) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, event messaging.Event), opts ...messaging.SubscriptionOption) error {
	options := &messaging.SubscriptionOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.Group != "" {
		// Consumer group subscription.
		groupName := options.Group
		consumerName := options.Name
		if consumerName == "" {
			consumerName = fmt.Sprintf("consumer_%s", topic)
		}

		// Create the consumer group; only new messages (after group creation) are processed.
		if err := r.createConsumerGroup(ctx, topic, groupName); err != nil {
			return err
		}

		// Start reading messages using consumer group semantics (XREADGroup with ">")
		go r.readMessagesConsumerGroup(ctx, topic, groupName, consumerName, handler)
		log.Printf("Subscribed to events of type: %s with consumer group: %s and consumer: %s", topic, groupName, consumerName)
	} else {
		// Non-consumer group subscription using XREAD.
		// Always use "$" to only get new messages.
		go r.readMessagesXRead(ctx, topic, handler)
		log.Printf("Subscribed to events of type: %s using XREAD", topic)
	}
	return nil
}

// createConsumerGroup creates a consumer group for the stream.
// It ignores the BUSYGROUP error if the group already exists.
func (r *RedisStream) createConsumerGroup(ctx context.Context, stream, groupName string) error {
	err := r.client.XGroupCreateMkStream(ctx, stream, groupName, "$").Err()
	if err != nil {
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			log.Printf("Redis Stream: consumer group %s already exists, continuing...", groupName)
			return nil
		}
		log.Printf("Redis Stream: error creating consumer group: %v", err)
		return err
	}
	return nil
}

// readMessagesConsumerGroup continuously reads messages from the stream using XREADGroup
// and invokes the handler. It uses ">" as the stream offset to fetch new messages.
func (r *RedisStream) readMessagesConsumerGroup(ctx context.Context, stream, group, consumer string, handler func(ctx context.Context, event messaging.Event)) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Consumer group subscription: stopping message reading for stream %s, consumer %s", stream, consumer)
			return
		default:
		}

		res, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Block:    0, // Block indefinitely until a message arrives.
		}).Result()
		if err != nil {
			log.Printf("Consumer group subscription: error reading from stream: %v", err)
			time.Sleep(errorSleepDuration)
			continue
		}
		r.processMessages(ctx, res, stream, group, handler)
	}
}

// readMessagesXRead continuously reads messages from the stream using XREAD
// and invokes the handler using a fixed start offset of "$" to process only new messages.
func (r *RedisStream) readMessagesXRead(ctx context.Context, stream string, handler func(ctx context.Context, event messaging.Event)) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("XREAD subscription: stopping message reading for stream %s", stream)
			return
		default:
		}

		res, err := r.client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{stream, "$"},
			Block:   0, // Block indefinitely until a message arrives.
		}).Result()
		if err != nil {
			log.Printf("XREAD subscription: error reading from stream: %v", err)
			time.Sleep(errorSleepDuration)
			continue
		}
		r.processMessagesXRead(ctx, res, stream, handler)
	}
}

// processMessages processes each message from consumer group subscriptions,
// calling the handler and acknowledging the message.
func (r *RedisStream) processMessages(ctx context.Context, streams []redis.XStream, stream, group string, handler func(ctx context.Context, event messaging.Event)) {
	for _, s := range streams {
		for _, message := range s.Messages {
			event, err := parseMessage(message)
			if err != nil {
				log.Printf("Consumer group subscription: error parsing message %v: %v", message.ID, err)
				continue
			}

			// Process the event concurrently.
			// For high volume, consider limiting concurrency.
			go handler(ctx, event)

			if _, ackErr := r.client.XAck(ctx, stream, group, message.ID).Result(); ackErr != nil {
				log.Printf("Consumer group subscription: failed to acknowledge message %v: %v", message.ID, ackErr)
			}
		}
	}
}

// processMessagesXRead processes each message from non-consumer group subscriptions
// and calls the handler.
func (r *RedisStream) processMessagesXRead(ctx context.Context, streams []redis.XStream, stream string, handler func(ctx context.Context, event messaging.Event)) {
	for _, s := range streams {
		for _, message := range s.Messages {
			event, err := parseMessage(message)
			if err != nil {
				log.Printf("XREAD subscription: error parsing message %v: %v", message.ID, err)
				continue
			}
			go handler(ctx, event)
		}
	}
}

// parseMessage unmarshals the event from the Redis stream message.
func parseMessage(message redis.XMessage) (messaging.Event, error) {
	eventData, ok := message.Values["event"].(string)
	if !ok {
		return messaging.Event{}, fmt.Errorf("invalid message format: missing 'event' field")
	}

	var event messaging.Event
	if err := json.Unmarshal([]byte(eventData), &event); err != nil {
		return messaging.Event{}, fmt.Errorf("failed to parse event data: %v", err)
	}

	return event, nil
}