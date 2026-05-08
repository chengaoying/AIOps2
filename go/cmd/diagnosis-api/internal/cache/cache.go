package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"aiops2/diagnosis-api/internal/engine"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr string, ttl time.Duration) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Cache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *Cache) Get(ctx context.Context, key string) (*engine.DiagnosisResult, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result engine.DiagnosisResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	result.UsedCache = true
	return &result, nil
}

func (c *Cache) Set(ctx context.Context, key string, result *engine.DiagnosisResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func BuildCacheKey(jobID string) string {
	return fmt.Sprintf("diagnosis:%s", jobID)
}
