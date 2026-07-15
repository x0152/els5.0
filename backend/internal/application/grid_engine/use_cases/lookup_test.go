package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/els/backend/internal/application/grid_engine/lookups"
	usecases "github.com/els/backend/internal/application/grid_engine/use_cases"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestLookupGrid_Forbidden(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, func(_ *iam.Actor) error { return shared.ErrForbidden })
	uc := usecases.NewLookupGridUseCase(cfg, lookups.NewResolver())

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.LookupGridQuery{})

	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestLookupGrid_NoResolver(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	uc := usecases.NewLookupGridUseCase(cfg, nil)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.LookupGridQuery{})

	test.ErrIs(t, err, shared.ErrUnavailable)
}

func TestLookupGrid_SourceNotInGrid(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	resolver := lookups.NewResolver(lookups.Source{ID: "accounts", Adapter: &stubAdapter{}})
	uc := usecases.NewLookupGridUseCase(cfg, resolver)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.LookupGridQuery{
		Queries: []usecases.LookupQueryRequest{{Source: grid.SourceID("missing")}},
	})

	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestLookupGrid_ValuesAndQMutuallyExclusive(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	resolver := lookups.NewResolver(lookups.Source{ID: "accounts", Adapter: &stubAdapter{}})
	uc := usecases.NewLookupGridUseCase(cfg, resolver)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.LookupGridQuery{
		Queries: []usecases.LookupQueryRequest{{
			Source: "accounts",
			Values: []string{"x"},
			Q:      "alice",
		}},
	})

	test.ErrIs(t, err, shared.ErrValidation)
}

func TestLookupGrid_OK_ResolveAndSearch(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	adapter := &stubAdapter{
		resolve: []lookups.Resolution{{Input: "o1", Key: "o1", Label: "Owner #1", MatchedBy: lookups.MatchByKey}},
		page:    lookups.Page{Items: []lookups.Item{{Key: "o2", Label: "Owner #2"}}},
	}
	resolver := lookups.NewResolver(lookups.Source{ID: "accounts", Adapter: adapter})
	uc := usecases.NewLookupGridUseCase(cfg, resolver)

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.LookupGridQuery{
		Queries: []usecases.LookupQueryRequest{
			{Source: "accounts", Values: []string{"o1"}},
			{Source: "accounts", Q: "owner"},
		},
	})

	test.NoErr(t, err)
	if len(res.Queries) != 2 {
		t.Fatalf("expected two query results, got %d", len(res.Queries))
	}
	if len(res.Queries[0].Resolutions) != 1 || res.Queries[0].Resolutions[0].Key != "o1" {
		t.Errorf("expected resolved value, got %+v", res.Queries[0].Resolutions)
	}
	if len(res.Queries[1].Items) != 1 || res.Queries[1].Items[0].Key != "o2" {
		t.Errorf("expected search items, got %+v", res.Queries[1].Items)
	}
}

func TestLookupGrid_AdapterFails(t *testing.T) {
	boom := errors.New("adapter blew up")
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	resolver := lookups.NewResolver(lookups.Source{ID: "accounts", Adapter: &stubAdapter{err: boom}})
	uc := usecases.NewLookupGridUseCase(cfg, resolver)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.LookupGridQuery{
		Queries: []usecases.LookupQueryRequest{{Source: "accounts", Q: "x"}},
	})

	if !errors.Is(err, boom) {
		t.Errorf("expected adapter error to propagate, got %v", err)
	}
}
