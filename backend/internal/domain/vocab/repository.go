package vocab

import "context"

type ListFilter struct {
	Search string
	Status Status
	Limit  int
	Offset int
}

type Repository interface {
	Create(ctx context.Context, unit Unit) (Unit, error)
	List(ctx context.Context, accountID string, filter ListFilter) ([]Unit, int, error)
	ExistsText(ctx context.Context, accountID, text string) (bool, error)
	UpdateStatus(ctx context.Context, accountID, id string, status Status) (Unit, error)
	Delete(ctx context.Context, accountID, id string) error
}
