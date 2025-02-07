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

// init loads configuration and sets up the default event bus.
func init() {
	var cfg Config
	if err := config.LoadConfig("application", []string{"."}, &cfg); err != nil {
		log.Fatalf("Redis Stream: error loading config: %v", err)
	}

	messaging.DefaultEventBus = NewRedisStream(&cfg.Messaging.RedisStream)
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

// Publish serializes only the event.Data field as JSON and adds the event to the stream.
func (r *RedisStream) Publish(ctx context.Context, event messaging.Event) error {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("Redis Stream: error serializing event data: %v", err)
		return err
	}

	_, err = r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: event.Type,
		Values: map[string]interface{}{
			"id":          event.ID,
			"source":      event.Source,
			"specversion": event.SpecVersion,
			"type":        event.Type,
			"data":        string(dataBytes),
			"time":        event.Time.Format(time.RFC3339),
		},
	}).Result()
	if err != nil {
		log.Printf("Redis Stream: failed to publish event: %v", err)
		return err
	}

	log.Printf("Published event: %v", event)
	return nil
}

// Subscribe creates (if needed) a consumer group and continuously reads messages from the stream.
// A shared background context is used within the loop.
func (r *RedisStream) Subscribe(eventType string, handler func(ctx context.Context, event messaging.Event)) error {
	groupName := fmt.Sprintf("consumer_group_%s", eventType)
	consumerName := fmt.Sprintf("consumer_%s", eventType)
	bg := context.Background()

	if err := r.client.XGroupCreateMkStream(bg, eventType, groupName, "$").Err(); err != nil && err != redis.Nil {
		log.Printf("Redis Stream: error creating consumer group: %v", err)
		return err
	}

	go func() {
		for {
			res, err := r.client.XReadGroup(bg, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{eventType, ">"},
				Block:    0,
			}).Result()
			if err != nil {
				log.Printf("Redis Stream: error reading from stream: %v", err)
				time.Sleep(errorSleepDuration)
				continue
			}

			for _, stream := range res {
				for _, message := range stream.Messages {
					event, parseErr := parseMessage(message)
					if parseErr != nil {
						log.Printf("Redis Stream: error parsing message %v: %v", message.ID, parseErr)
						continue
					}

					go handler(bg, event)

					if _, ackErr := r.client.XAck(bg, eventType, groupName, message.ID).Result(); ackErr != nil {
						log.Printf("Redis Stream: failed to acknowledge message %v: %v", message.ID, ackErr)
					}
				}
			}
		}
	}()

	log.Printf("Subscribed to events of type: %s", eventType)
	return nil
}

// parseMessage extracts and validates fields from a Redis stream message and returns an Event.
func parseMessage(message redis.XMessage) (messaging.Event, error) {
	id, idOk := message.Values["id"].(string)
	source, sourceOk := message.Values["source"].(string)
	specVersion, specVersionOk := message.Values["specversion"].(string)
	eventType, typeOk := message.Values["type"].(string)
	timeStr, timeOk := message.Values["time"].(string)
	dataJSON, dataOk := message.Values["data"].(string)

	if !idOk || !sourceOk || !specVersionOk || !typeOk || !timeOk || !dataOk {
		return messaging.Event{}, fmt.Errorf("invalid message format")
	}

	eventTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return messaging.Event{}, fmt.Errorf("invalid time format: %v", err)
	}

	var eventData map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &eventData); err != nil {
		return messaging.Event{}, fmt.Errorf("failed to parse event data: %v", err)
	}

	return messaging.Event{
		ID:          id,
		Source:      source,
		SpecVersion: specVersion,
		Type:        eventType,
		Data:        eventData,
		Time:        eventTime,
	}, nil
}
