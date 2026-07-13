package gridspec

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"github.com/els/backend/internal/application/grid_engine/lookups"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/database"
)

type Page struct {
	Limit  int32
	Offset int32
}

type CRUD[E any] struct {
	List        func(ctx context.Context, actor *iam.Actor, page Page) (items []E, total int64, err error)
	GetByID     func(ctx context.Context, actor *iam.Actor, id string) (E, error)
	Create      func(ctx context.Context, actor *iam.Actor, data map[grid.ColumnID]any) (E, error)
	Update      func(ctx context.Context, e E) error
	AfterUpdate func(ctx context.Context, actor *iam.Actor, before, after grid.Row, e E) error
	Delete      func(ctx context.Context, actor *iam.Actor, id string) error
	Version     func(e E) int64
}

type Config[E any] struct {
	BasePath string
	Tag      string
	Summary  string

	Authorize func(actor *iam.Actor) error
	Grid      func(actor *iam.Actor) grid.Grid[E]
	CRUD      CRUD[E]

	DefaultLimit int32
	MaxLimit     int32
}

func (c Config[E]) Normalize(limit, offset int32) Page {
	def, max := c.limits()
	if limit <= 0 {
		limit = def
	}
	if limit > max {
		limit = max
	}
	if offset < 0 {
		offset = 0
	}
	return Page{Limit: limit, Offset: offset}
}

func (c Config[E]) limits() (int32, int32) {
	def := c.DefaultLimit
	max := c.MaxLimit
	if def <= 0 {
		def = 50
	}
	if max <= 0 {
		max = 500
	}
	return def, max
}

type Binding struct {
	BasePath string
	Tag      string
	Summary  string
	Register func(api huma.API, auth *authx.Authenticator, resolver *lookups.Resolver, tx database.TxRunner)
}
