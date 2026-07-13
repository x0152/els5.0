package ports

import (
	"context"
	"time"
)

type LoginAttemptStore interface {
	IsLocked(ctx context.Context, accountID string) (bool, error)
	Fail(ctx context.Context, accountID string, threshold int, window time.Duration) error
	Reset(ctx context.Context, accountID string) error
}
