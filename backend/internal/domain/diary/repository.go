package diary

import (
	"context"
	"time"
)

type Repository interface {
	Insert(ctx context.Context, e Entry) error
	UpdateReply(ctx context.Context, e Entry) error
	GetByDate(ctx context.Context, accountID string, date time.Time) (Entry, error)
	List(ctx context.Context, accountID string, limit, offset int32) ([]Entry, int64, error)
	Latest(ctx context.Context, accountID string, n int32) ([]Entry, error)
	Dates(ctx context.Context, accountID string, limit int32) ([]time.Time, error)
	DeleteAll(ctx context.Context, accountID string) error
}
