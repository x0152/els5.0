package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr           string
	Password       string
	DB             int
	ConnectTimeout time.Duration
}

func New(ctx context.Context, cfg Config) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("addr is empty")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return client, nil
}
