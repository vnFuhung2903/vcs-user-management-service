package interfaces

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type IRedisClient interface {
	Del(ctx context.Context, key string) error
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(client *redis.Client) IRedisClient {
	return &RedisClient{client: client}
}

func (c *RedisClient) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
