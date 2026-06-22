// Package redisx centralizes Redis client setup.
package redisx

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Connect parses a redis:// URL, opens a client, and verifies connectivity.
func Connect(ctx context.Context, url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
