package usecases_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/els/backend/internal/application/grid_engine/gridspec"
	"github.com/els/backend/internal/application/grid_engine/lookups"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type fakeEntity struct {
	id      string
	name    string
	owner   string
	version int64
}

type fakeStore struct {
	mu      sync.Mutex
	items   map[string]*fakeEntity
	nextID  atomic.Int64
	listErr error
	getErr  error
	createErr error
	updateErr error
	deleteErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{items: map[string]*fakeEntity{}}
}

func (s *fakeStore) snapshot() []*fakeEntity {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*fakeEntity, 0, len(s.items))
	for _, e := range s.items {
		cp := *e
		out = append(out, &cp)
	}
	return out
}

func (s *fakeStore) put(e *fakeEntity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *e
	s.items[e.id] = &cp
}

func sampleGrid() grid.Grid[*fakeEntity] {
	return grid.New(grid.Spec[*fakeEntity]{
		Columns: []grid.Column{
			{ID: "name", Type: grid.TypeText, Required: true},
			{ID: "owner", Type: grid.TypeRef, Ref: &grid.RefSpec{Source: "accounts", KeyField: "id", LabelField: "name"}},
		},
		Row: func(e *fakeEntity) grid.Row {
			return grid.Row{
				ID:          e.id,
				BaseVersion: e.version,
				Cells: map[grid.ColumnID]any{
					"name":  e.name,
					"owner": e.owner,
				},
			}
		},
		ApplyPatch: func(e *fakeEntity, data map[grid.ColumnID]any) error {
			if v, ok := data["name"]; ok {
				if s, ok := v.(string); ok {
					if s == "" {
						return shared.Validation(errors.New("name: must not be empty"))
					}
					e.name = s
				}
			}
			if v, ok := data["owner"]; ok {
				if s, ok := v.(string); ok {
					e.owner = s
				}
			}
			return nil
		},
	})
}

func cfgFor(store *fakeStore, authorize func(*iam.Actor) error) gridspec.Config[*fakeEntity] {
	return gridspec.Config[*fakeEntity]{
		Authorize: authorize,
		Grid:      func(_ *iam.Actor) grid.Grid[*fakeEntity] { return sampleGrid() },
		CRUD: gridspec.CRUD[*fakeEntity]{
			List: func(_ context.Context, _ *iam.Actor, page gridspec.Page) ([]*fakeEntity, int64, error) {
				if store.listErr != nil {
					return nil, 0, store.listErr
				}
				items := store.snapshot()
				return items, int64(len(items)), nil
			},
			GetByID: func(_ context.Context, _ *iam.Actor, id string) (*fakeEntity, error) {
				if store.getErr != nil {
					return nil, store.getErr
				}
				store.mu.Lock()
				defer store.mu.Unlock()
				e, ok := store.items[id]
				if !ok {
					return nil, shared.ErrNotFound
				}
				cp := *e
				return &cp, nil
			},
			Create: func(_ context.Context, _ *iam.Actor, data map[grid.ColumnID]any) (*fakeEntity, error) {
				if store.createErr != nil {
					return nil, store.createErr
				}
				e := &fakeEntity{
					id:      string(rune(int('A') + int(store.nextID.Add(1)))),
					version: 1,
				}
				if v, ok := data["name"].(string); ok {
					e.name = v
				}
				if v, ok := data["owner"].(string); ok {
					e.owner = v
				}
				store.put(e)
				return e, nil
			},
			Update: func(_ context.Context, e *fakeEntity) error {
				if store.updateErr != nil {
					return store.updateErr
				}
				store.mu.Lock()
				defer store.mu.Unlock()
				cur, ok := store.items[e.id]
				if !ok {
					return shared.ErrNotFound
				}
				cur.name = e.name
				cur.owner = e.owner
				cur.version++
				return nil
			},
			Delete: func(_ context.Context, _ *iam.Actor, id string) error {
				if store.deleteErr != nil {
					return store.deleteErr
				}
				store.mu.Lock()
				defer store.mu.Unlock()
				delete(store.items, id)
				return nil
			},
			Version: func(e *fakeEntity) int64 { return e.version },
		},
	}
}

type stubAdapter struct {
	hydrate map[string]string
	resolve []lookups.Resolution
	unres   []string
	page    lookups.Page
	err     error
}

func (s *stubAdapter) Hydrate(_ context.Context, _ *iam.Actor, _ []string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.hydrate, nil
}

func (s *stubAdapter) Resolve(_ context.Context, _ *iam.Actor, _ []string) ([]lookups.Resolution, []string, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.resolve, s.unres, nil
}

func (s *stubAdapter) Search(_ context.Context, _ *iam.Actor, _ string, _ int32, _ string) (lookups.Page, error) {
	if s.err != nil {
		return lookups.Page{}, s.err
	}
	return s.page, nil
}
