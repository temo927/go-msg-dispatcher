package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr, password string, db int, ttl time.Duration) *Cache {
	return &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ttl: ttl,
	}
}

func (c *Cache) SetSentMeta(ctx context.Context, msgID string, meta map[string]string) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "msg:"+msgID+":meta", data, c.ttl).Err()
}
