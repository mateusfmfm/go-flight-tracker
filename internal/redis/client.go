package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"go-flight-tracker/internal/config"
	"go-flight-tracker/internal/flight"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb     *redis.Client
	channel string
}

func NewClient(ctx context.Context, cfg config.Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	//Validate Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", cfg.RedisAddr, err)
	}
	log.Println("Successfully connected to Redis!")

	return &Client{
		rdb:     rdb,
		channel: cfg.RedisChannel,
	}, nil
}

// SetCache holds a snapshot of the aircrafts in cache with expiration time (TTL)
func (c *Client) SetCache(ctx context.Context, key string, aircrafts []*flight.Aircraft, ttl time.Duration) error {
	data, err := json.Marshal(aircrafts)
	if err != nil {
		return fmt.Errorf("failed to marshal aircrafts for cache: %w", err)
	}

	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// PublishAircrafts publish at Pub/Sub channel
func (c *Client) PublishAircrafts(ctx context.Context, aircrafts []*flight.Aircraft) error {
	data, err := json.Marshal(aircrafts)
	if err != nil {
		return fmt.Errorf("failed to marshal aircrafts for redis: %w", err)
	}

	return c.rdb.Publish(ctx, c.channel, data).Err()
}

// SubscribeAircrafts listen updates of Pub/Sub channel and deliver in a Go Channel
func (c *Client) SubscribeAircrafts(ctx context.Context) (<-chan []*flight.Aircraft, error) {
	pubsub := c.rdb.Subscribe(ctx, c.channel)

	//Guarantee subscription is established
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("failed to subscribe to redis channel %s: %w", c.channel, err)
	}

	out := make(chan []*flight.Aircraft)
	go func() {
		defer close(out)
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				_ = pubsub.Close()
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var aircrafts []*flight.Aircraft
				if err := json.Unmarshal([]byte(msg.Payload), &aircrafts); err != nil {
					log.Printf("failed to unmarshal aircrafts from redis: %v", err)
					continue
				}
				out <- aircrafts
			}
		}
	}()

	return out, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
