package redisstream

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ebrickdev/ebrick/config"
	"github.com/ebrickdev/ebrick/messaging"
	"github.com/redis/go-redis/v9"
)

func init() {
	var cfg Config
	err := config.LoadConfig("application", []string{"."}, &cfg)
	if err != nil {
		log.Fatalf("Redis Stream: error loading config: %v", err)
	}

	messaging.DefaultEventBus = NewRedisStream(&cfg.Messaging.RedisStream)
}

type RedisStream struct {
	client *redis.Client
	stream string
}

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
	log.Println("Redis Stream initialized successfully")

	return &RedisStream{
		client: client,
		stream: cfg.StreamName,
	}
}

func (r *RedisStream) Close() error {
	log.Println("Closing Redis connection")
	return r.client.Close()
}

func (r *RedisStream) Publish(ctx context.Context, event messaging.Event) error {
	_, err := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.stream,
		Values: map[string]interface{}{
			"id":      event.ID,
			"type":    event.Type,
			"payload": string(event.Payload),
		},
	}).Result()

	if err != nil {
		log.Printf("Redis Stream: failed to publish event: %v", err)
		return err
	}

	log.Printf("Published event: %v", event)
	return nil
}

func (r *RedisStream) Subscribe(eventType string, handler func(ctx context.Context, event messaging.Event)) error {
	groupName := fmt.Sprintf("consumer_group_%s", eventType)
	consumerName := fmt.Sprintf("consumer_%s", eventType)

	if err := r.client.XGroupCreateMkStream(context.Background(), r.stream, groupName, "$").Err(); err != nil && err != redis.Nil {
		log.Printf("Redis Stream: error creating consumer group: %v", err)
		return err
	}

	go func() {
		for {
			res, err := r.client.XReadGroup(context.Background(), &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{r.stream, ">"},
				Block:    0,
			}).Result()

			if err != nil {
				log.Printf("Redis Stream: error reading from stream: %v", err)
				time.Sleep(time.Second)
				continue
			}

			for _, stream := range res {
				for _, message := range stream.Messages {
					event, parseErr := parseMessage(message)
					if parseErr != nil {
						log.Printf("Redis Stream: error parsing message %v: %v", message.ID, parseErr)
						continue
					}

					go handler(context.Background(), event)

					if _, ackErr := r.client.XAck(context.Background(), r.stream, groupName, message.ID).Result(); ackErr != nil {
						log.Printf("Redis Stream: failed to acknowledge message %v: %v", message.ID, ackErr)
					}
				}
			}
		}
	}()

	log.Printf("Subscribed to events of type: %s", eventType)
	return nil
}

// parseMessage safely parses Redis message values into an event.
func parseMessage(message redis.XMessage) (messaging.Event, error) {
	id, idOk := message.Values["id"].(string)
	eventType, typeOk := message.Values["type"].(string)
	payload, payloadOk := message.Values["payload"].(string)

	if !idOk || !typeOk || !payloadOk {
		return messaging.Event{}, fmt.Errorf("invalid message format")
	}

	return messaging.Event{
		ID:      id,
		Type:    eventType,
		Payload: []byte(payload),
	}, nil
}
