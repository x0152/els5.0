package vocab

import (
	"context"
	"time"
)

type ListFilter struct {
	Search string
	Status Status
	Limit  int
	Offset int
}

type Repository interface {
	Create(ctx context.Context, unit Unit) (Unit, error)
	Get(ctx context.Context, accountID, id string) (Unit, error)
	List(ctx context.Context, accountID string, filter ListFilter) ([]Unit, int, error)
	ExistsText(ctx context.Context, accountID, text, kind string) (bool, error)
	UpdateDetails(ctx context.Context, unit Unit) error
	UpdateStatus(ctx context.Context, accountID, id string, status Status) (Unit, error)
	UpdateProgress(ctx context.Context, unit Unit) (Unit, error)
	CountDue(ctx context.Context, accountID string, since time.Time) (int, error)
	Delete(ctx context.Context, accountID, id string) error
}
