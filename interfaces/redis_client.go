package interfaces

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type IRedisClient interface {
	Del(ctx context.Context, key string) error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(client *redis.Client) IRedisClient {
	return &redisClient{client: client}
}

func (c *redisClient) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
