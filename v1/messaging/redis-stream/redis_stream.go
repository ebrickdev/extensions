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

const (
	errorSleepDuration = time.Second
)

// Init loads configuration and sets up the default event bus.
func Init() *RedisStream {
	var cfg Config
	if err := config.LoadConfig("application", []string{"."}, &cfg); err != nil {
		log.Fatalf("Redis Stream: error loading config: %v", err)
	}

	return NewRedisStream(&cfg.Messaging.RedisStream)
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

// Publish serializes the entire event as JSON and adds it to the stream.
func (r *RedisStream) Publish(ctx context.Context, event messaging.Event) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Redis Stream: failed to serialize event: %v", err)
		return err
	}

	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: event.Type,
		Values: map[string]interface{}{
			"event": string(eventData),
		},
	}).Result()
	if err != nil {
		log.Printf("Redis Stream: failed to publish event: %v", err)
		return err
	}

	// log.Printf("Published event: %v", string(eventData))
	return nil
}

func (r *RedisStream) Subscribe(eventType string, handler func(ctx context.Context, event messaging.Event)) error {
	groupName := fmt.Sprintf("consumer_group_%s", eventType)
	consumerName := fmt.Sprintf("consumer_%s", eventType)
	bg := context.Background()

	if err := r.createConsumerGroup(bg, eventType, groupName); err != nil {
		return err
	}

	go r.readMessages(bg, eventType, groupName, consumerName, handler)

	log.Printf("Subscribed to events of type: %s", eventType)
	return nil
}

// createConsumerGroup handles consumer group creation and BUSYGROUP errors.
func (r *RedisStream) createConsumerGroup(ctx context.Context, stream, groupName string) error {
	err := r.client.XGroupCreateMkStream(ctx, stream, groupName, "$").Err()
	if err != nil {
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			log.Printf("Redis Stream: consumer group already exists, continuing...")
			return nil
		} else if err != redis.Nil {
			log.Printf("Redis Stream: error creating consumer group: %v", err)
			return err
		}
	}
	return nil
}

// readMessages continuously reads messages from the stream and invokes the handler.
func (r *RedisStream) readMessages(ctx context.Context, stream, group, consumer string, handler func(ctx context.Context, event messaging.Event)) {
	for {
		res, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Block:    0,
		}).Result()
		if err != nil {
			log.Printf("Redis Stream: error reading from stream: %v", err)
			time.Sleep(errorSleepDuration)
			continue
		}
		r.processMessages(ctx, res, stream, group, handler)
	}
}

// processMessages processes each message, calling the handler and acknowledging the message.
func (r *RedisStream) processMessages(ctx context.Context, streams []redis.XStream, stream, group string, handler func(ctx context.Context, event messaging.Event)) {
	for _, s := range streams {
		for _, message := range s.Messages {
			event, err := parseMessage(message)
			if err != nil {
				log.Printf("Redis Stream: error parsing message %v: %v", message.ID, err)
				continue
			}

			go handler(ctx, event)

			if _, ackErr := r.client.XAck(ctx, stream, group, message.ID).Result(); ackErr != nil {
				log.Printf("Redis Stream: failed to acknowledge message %v: %v", message.ID, ackErr)
			}
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
