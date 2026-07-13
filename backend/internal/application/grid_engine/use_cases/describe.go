package usecases

import (
	"context"
	"time"

	"github.com/els/backend/internal/application/grid_engine/gridspec"
	"github.com/els/backend/internal/application/grid_engine/lookups"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
)

type DescribeGridUseCase[E any] struct {
	cfg     gridspec.Config[E]
	lookups *lookups.Resolver
}

func NewDescribeGridUseCase[E any](cfg gridspec.Config[E], resolver *lookups.Resolver) *DescribeGridUseCase[E] {
	return &DescribeGridUseCase[E]{cfg: cfg, lookups: resolver}
}

type DescribeGridQuery struct {
	Limit  int32
	Offset int32
}

func (uc *DescribeGridUseCase[E]) Execute(ctx context.Context, actor *iam.Actor, q DescribeGridQuery) (DescribeResult, error) {
	// 1. Check the actor's access to this grid.
	if uc.cfg.Authorize != nil {
		if err := uc.cfg.Authorize(actor); err != nil {
			return DescribeResult{}, err
		}
	}

	// 2. Take the grid for the actor and normalize pagination.
	g := uc.cfg.Grid(actor)
	page := uc.cfg.Normalize(q.Limit, q.Offset)

	// 3. Load the entity page + total.
	items, total, err := uc.cfg.CRUD.List(ctx, actor, page)
	if err != nil {
		return DescribeResult{}, err
	}

	// 4. Map entities to table rows.
	rows := make([]grid.Row, 0, len(items))
	for _, item := range items {
		rows = append(rows, g.RowOf(item))
	}

	// 5. Load human-readable labels for ref columns in one batch call.
	hydrated := map[grid.SourceID]map[string]string{}
	if uc.lookups != nil {
		hydrated, err = uc.lookups.HydrateRefs(ctx, actor, g.Schema(), rows)
		if err != nil {
			return DescribeResult{}, err
		}
	}

	// 6. Build the response with schema, rows, and source metadata.
	return DescribeResult{
		SchemaVersion: g.Version(),
		Columns:       g.Schema().Columns,
		Sources:       g.Sources(),
		Rows:          rows,
		Total:         total,
		Limit:         page.Limit,
		Offset:        page.Offset,
		RefsHydrated:  hydrated,
		GeneratedAt:   time.Now().UTC(),
	}, nil
}
