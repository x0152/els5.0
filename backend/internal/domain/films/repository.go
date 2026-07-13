package films

import (
	"context"
	"time"
)

type Repository interface {
	List(ctx context.Context) ([]Film, error)
	Get(ctx context.Context, id string) (Film, error)
	Create(ctx context.Context, film Film) error
	Update(ctx context.Context, film Film) error
	Delete(ctx context.Context, id string) error

	ListSeries(ctx context.Context) ([]Series, error)
	GetSeries(ctx context.Context, title string) (Series, error)
	UpsertSeries(ctx context.Context, series Series) error
	RenameSeries(ctx context.Context, oldTitle, newTitle string) error

	SaveProgress(ctx context.Context, ownerID, filmID string, positionMs int, updatedAt time.Time) error
	ListProgress(ctx context.Context, ownerID string) (map[string]int, error)
}
