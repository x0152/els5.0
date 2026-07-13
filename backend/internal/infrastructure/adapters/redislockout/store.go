package redislockout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/domain/shared/ports"
)

const defaultPrefix = "auth:fail:"

type Store struct {
	client *redis.Client
	prefix string
}

var _ ports.LoginAttemptStore = (*Store)(nil)

func NewStore(client *redis.Client, prefix string) *Store {
	if prefix == "" {
		prefix = defaultPrefix
	}
	return &Store{client: client, prefix: prefix}
}

func (s *Store) IsLocked(ctx context.Context, accountID string) (bool, error) {
	if accountID == "" {
		return false, nil
	}
	exists, err := s.client.Exists(ctx, s.lockKey(accountID)).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists: %w", err)
	}
	return exists > 0, nil
}

func (s *Store) Fail(ctx context.Context, accountID string, threshold int, window time.Duration) error {
	if accountID == "" || threshold <= 0 || window <= 0 {
		return nil
	}
	failKey := s.failKey(accountID)
	pipe := s.client.TxPipeline()
	incr := pipe.Incr(ctx, failKey)
	pipe.Expire(ctx, failKey, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis incr: %w", err)
	}
	if incr.Val() < int64(threshold) {
		return nil
	}
	if err := s.client.Set(ctx, s.lockKey(accountID), "1", window).Err(); err != nil {
		return fmt.Errorf("redis set lock: %w", err)
	}
	return nil
}

func (s *Store) Reset(ctx context.Context, accountID string) error {
	if accountID == "" {
		return nil
	}
	if err := s.client.Del(ctx, s.failKey(accountID), s.lockKey(accountID)).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func (s *Store) failKey(accountID string) string { return s.prefix + "count:" + accountID }
func (s *Store) lockKey(accountID string) string { return s.prefix + "lock:" + accountID }
