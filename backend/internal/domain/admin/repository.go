package admin

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
)

type Repository interface {
	Create(ctx context.Context, a *Administrator) error
	Update(ctx context.Context, a *Administrator) error
	Delete(ctx context.Context, id ID) error
	GetByID(ctx context.Context, id ID) (*Administrator, error)
	GetByAccountID(ctx context.Context, id iam.AccountID) (*Administrator, error)
	List(ctx context.Context, filter Filter, limit, offset int32) ([]*Administrator, int64, error)
	Count(ctx context.Context) (int64, error)
}
