package redisratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	client *redis.Client
	prefix string
}

func New(client *redis.Client, prefix string) *Limiter {
	if prefix == "" {
		prefix = "ratelimit:"
	}
	return &Limiter{client: client, prefix: prefix}
}

func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 || window <= 0 {
		return false, errors.New("rateimit: limit and window must be > 0")
	}
	if l.client == nil {
		return false, errors.New("rateimit: redis client is nil")
	}
	full := l.prefix + key
	pipe := l.client.TxPipeline()
	incr := pipe.Incr(ctx, full)
	pipe.Expire(ctx, full, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("redis ratelimit: %w", err)
	}
	if incr.Val() > int64(limit) {
		return false, nil
	}
	return true, nil
}
