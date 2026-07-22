package studio

import "context"

type Repository interface {
	ListAreas(ctx context.Context, accountID string) ([]AreaStats, error)
	InsertArea(ctx context.Context, a Area) error
	DeleteArea(ctx context.Context, accountID, id string) error
	ListItems(ctx context.Context, accountID, areaID string) ([]Item, error)
	Get(ctx context.Context, accountID, id string) (Item, error)
	Insert(ctx context.Context, i Item) error
	Update(ctx context.Context, i Item) error
	Delete(ctx context.Context, accountID, id string) error
}
