package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/application/grid_engine/gridspec"
	"github.com/els/backend/internal/application/grid_engine/lookups"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type LookupGridUseCase[E any] struct {
	cfg     gridspec.Config[E]
	lookups *lookups.Resolver
}

func NewLookupGridUseCase[E any](cfg gridspec.Config[E], resolver *lookups.Resolver) *LookupGridUseCase[E] {
	return &LookupGridUseCase[E]{cfg: cfg, lookups: resolver}
}

type LookupGridQuery struct {
	Queries []LookupQueryRequest
}

func (uc *LookupGridUseCase[E]) Execute(ctx context.Context, actor *iam.Actor, q LookupGridQuery) (LookupResult, error) {
	// 1. Check the actor's access to the grid and that lookups are configured.
	if uc.cfg.Authorize != nil {
		if err := uc.cfg.Authorize(actor); err != nil {
			return LookupResult{}, err
		}
	}
	if uc.lookups == nil {
		return LookupResult{}, fmt.Errorf("%w: lookups are not configured", shared.ErrUnavailable)
	}

	// 2. Prepare the grid and response accumulator.
	g := uc.cfg.Grid(actor)
	out := LookupResult{Queries: make([]LookupQueryResult, 0, len(q.Queries))}

	// 3. Process requests for each source in order.
	for _, req := range q.Queries {

		if !g.HasSource(req.Source) {
			return LookupResult{}, fmt.Errorf("%w: source %q is not allowed for this grid", shared.ErrForbidden, req.Source)
		}
		adapter, err := uc.lookups.Adapter(req.Source)
		if err != nil {
			return LookupResult{}, err
		}

		hasValues := len(req.Values) > 0
		hasQuery := req.Q != ""
		if hasValues && hasQuery {
			return LookupResult{}, shared.Validation(fmt.Errorf("lookup.query: only one of `values` or `q` may be provided"))
		}

		if hasValues {
			items, unresolved, err := adapter.Resolve(ctx, actor, req.Values)
			if err != nil {
				return LookupResult{}, err
			}
			out.Queries = append(out.Queries, LookupQueryResult{
				Source:      req.Source,
				Resolutions: items,
				Unresolved:  unresolved,
			})
			continue
		}

		page, err := adapter.Search(ctx, actor, req.Q, req.Limit, req.Cursor)
		if err != nil {
			return LookupResult{}, err
		}
		out.Queries = append(out.Queries, LookupQueryResult{
			Source:     req.Source,
			Items:      page.Items,
			NextCursor: page.NextCursor,
		})
	}
	return out, nil
}
