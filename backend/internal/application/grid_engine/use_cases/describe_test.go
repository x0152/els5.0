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

func TestDescribeGrid_Forbidden(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, func(_ *iam.Actor) error { return shared.ErrForbidden })
	uc := usecases.NewDescribeGridUseCase(cfg, nil)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.DescribeGridQuery{})

	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestDescribeGrid_OK_RowsAndPagination(t *testing.T) {
	store := newFakeStore()
	store.put(&fakeEntity{id: "a1", name: "Alice", owner: "o1", version: 1})
	store.put(&fakeEntity{id: "a2", name: "Bob", owner: "o2", version: 1})
	cfg := cfgFor(store, nil)
	cfg.DefaultLimit = 25
	uc := usecases.NewDescribeGridUseCase(cfg, nil)

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.DescribeGridQuery{Limit: 0, Offset: -3})

	test.NoErr(t, err)
	if len(res.Rows) != 2 || res.Total != 2 {
		t.Errorf("expected 2 rows, got %d (total=%d)", len(res.Rows), res.Total)
	}
	if res.Limit != 25 || res.Offset != 0 {
		t.Errorf("expected normalized page (25,0), got (%d,%d)", res.Limit, res.Offset)
	}
}

func TestDescribeGrid_HydratesRefs(t *testing.T) {
	store := newFakeStore()
	store.put(&fakeEntity{id: "a1", name: "Alice", owner: "o1", version: 1})
	cfg := cfgFor(store, nil)

	resolver := lookups.NewResolver(lookups.Source{
		ID:      "accounts",
		Adapter: &stubAdapter{hydrate: map[string]string{"o1": "Owner #1"}},
	})
	uc := usecases.NewDescribeGridUseCase(cfg, resolver)

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.DescribeGridQuery{Limit: 10})

	test.NoErr(t, err)
	if got := res.RefsHydrated["accounts"]["o1"]; got != "Owner #1" {
		t.Errorf("expected hydrated label, got %q", got)
	}
}

func TestDescribeGrid_RepoFails(t *testing.T) {
	boom := errors.New("list failed")
	store := newFakeStore()
	store.listErr = boom
	cfg := cfgFor(store, nil)
	uc := usecases.NewDescribeGridUseCase(cfg, nil)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.DescribeGridQuery{Limit: 10})

	if !errors.Is(err, boom) {
		t.Errorf("expected list error to propagate, got %v", err)
	}
}

func TestDescribeGrid_PaginationCappedAtMax(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	cfg.DefaultLimit = 25
	cfg.MaxLimit = 100
	uc := usecases.NewDescribeGridUseCase(cfg, nil)

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.DescribeGridQuery{Limit: 9999})

	test.NoErr(t, err)
	if res.Limit != 100 {
		t.Errorf("expected limit capped at 100, got %d", res.Limit)
	}
}

func TestDescribeGrid_ColumnsAndSchemaVersion(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	uc := usecases.NewDescribeGridUseCase(cfg, nil)

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.DescribeGridQuery{Limit: 10})

	test.NoErr(t, err)
	g := sampleGrid()
	if res.SchemaVersion != g.Version() {
		t.Errorf("expected schema version %s, got %s", g.Version(), res.SchemaVersion)
	}
	if len(res.Columns) != len(g.Schema().Columns) {
		t.Errorf("expected %d columns, got %d", len(g.Schema().Columns), len(res.Columns))
	}
	if len(res.Sources) == 0 || res.Sources[0] != grid.SourceID("accounts") {
		t.Errorf("expected accounts source, got %v", res.Sources)
	}
}
