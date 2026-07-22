package onboarding

import (
	"context"
	"time"
)

type Repository interface {
	Watermarks(ctx context.Context, accountID string) (map[string]int, error)
	SaveWatermarks(ctx context.Context, accountID string, values map[string]int, now time.Time) error
	LiveCounts(ctx context.Context, accountID string) (map[string]int, error)
	Acks(ctx context.Context, accountID string) (map[string]bool, error)
	SaveAcks(ctx context.Context, accountID string, itemIDs []string, now time.Time) error
}
