package rediscache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func NewClient(addr string) *Client {
	return &Client{Client: redis.NewClient(&redis.Options{Addr: addr})}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
