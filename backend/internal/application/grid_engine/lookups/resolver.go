package lookups

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type Resolver struct {
	adapters map[grid.SourceID]Adapter
}

func NewResolver(sources ...Source) *Resolver {
	m := make(map[grid.SourceID]Adapter, len(sources))
	for _, s := range sources {
		if s.Adapter == nil {
			panic(fmt.Sprintf("grid_engine/lookups: nil adapter for source %q", s.ID))
		}
		if _, exists := m[s.ID]; exists {
			panic(fmt.Sprintf("grid_engine/lookups: duplicate source %q", s.ID))
		}
		m[s.ID] = s.Adapter
	}
	return &Resolver{adapters: m}
}

func (r *Resolver) Adapter(src grid.SourceID) (Adapter, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: lookups resolver not configured", shared.ErrUnavailable)
	}
	a, ok := r.adapters[src]
	if !ok {
		return nil, fmt.Errorf("%w: lookup source %q is not registered", shared.ErrNotFound, src)
	}
	return a, nil
}

func (r *Resolver) Hydrate(ctx context.Context, actor *iam.Actor, src grid.SourceID, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	a, err := r.Adapter(src)
	if err != nil {
		return nil, err
	}
	return a.Hydrate(ctx, actor, dedup(keys))
}

func (r *Resolver) HydrateRefs(ctx context.Context, actor *iam.Actor, schema grid.Schema, rows []grid.Row) (map[grid.SourceID]map[string]string, error) {
	if len(rows) == 0 {
		return map[grid.SourceID]map[string]string{}, nil
	}
	keysBySource := map[grid.SourceID]map[string]struct{}{}
	for _, c := range schema.Columns {
		if !c.IsRef() {
			continue
		}
		for _, row := range rows {
			v, ok := row.Cells[c.ID]
			if !ok || v == nil {
				continue
			}
			for _, k := range collectKeys(v) {
				if k == "" {
					continue
				}
				if _, ok := keysBySource[c.Ref.Source]; !ok {
					keysBySource[c.Ref.Source] = map[string]struct{}{}
				}
				keysBySource[c.Ref.Source][k] = struct{}{}
			}
		}
	}
	out := map[grid.SourceID]map[string]string{}
	for src, set := range keysBySource {
		keys := make([]string, 0, len(set))
		for k := range set {
			keys = append(keys, k)
		}
		labels, err := r.Hydrate(ctx, actor, src, keys)
		if err != nil {
			return nil, err
		}
		out[src] = labels
	}
	return out, nil
}

func collectKeys(v any) []string {
	switch val := v.(type) {
	case string:
		return []string{val}
	case []string:
		return val
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func dedup(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
