package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	usecases "github.com/els/backend/internal/application/grid_engine/use_cases"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/database"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestApplyGrid_Forbidden(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, func(_ *iam.Actor) error { return shared.ErrForbidden })
	uc := usecases.NewApplyGridUseCase(cfg, database.Noop())

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{SchemaVersion: "v1"})

	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestApplyGrid_SchemaVersionMismatch(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	uc := usecases.NewApplyGridUseCase(cfg, database.Noop())

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{SchemaVersion: "wrong"})

	test.ErrIs(t, err, shared.ErrConflict)
}

func TestApplyGrid_OK_CreateAndUpdate(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	uc := usecases.NewApplyGridUseCase(cfg, database.Noop())
	g := sampleGrid()

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{
		SchemaVersion: g.Version(),
		Operations: []grid.Op{
			{Kind: grid.OpCreate, TempID: "t1", Data: map[grid.ColumnID]any{"name": "Alice"}},
		},
	})

	test.NoErr(t, err)
	if len(res.Applied) != 1 || res.Applied[0].TempID != "t1" {
		t.Fatalf("expected one applied op for tempID=t1, got %+v", res.Applied)
	}
	if len(res.Failed) != 0 {
		t.Errorf("expected no failed ops, got %+v", res.Failed)
	}
}

func TestApplyGrid_AfterUpdateReceivesBeforeAndAfterRows(t *testing.T) {
	store := newFakeStore()
	store.put(&fakeEntity{id: "x1", name: "A", version: 1})
	cfg := cfgFor(store, nil)
	var calls []struct {
		before string
		after  string
	}
	cfg.CRUD.AfterUpdate = func(_ context.Context, _ *iam.Actor, before, after grid.Row, _ *fakeEntity) error {
		calls = append(calls, struct {
			before string
			after  string
		}{
			before: before.Cells["name"].(string),
			after:  after.Cells["name"].(string),
		})
		return nil
	}
	uc := usecases.NewApplyGridUseCase(cfg, database.Noop())
	g := sampleGrid()

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{
		SchemaVersion: g.Version(),
		Operations: []grid.Op{
			{Kind: grid.OpUpdate, ID: "x1", BaseVersion: 1, Data: map[grid.ColumnID]any{"name": "B"}},
		},
	})

	test.NoErr(t, err)
	if len(res.Failed) != 0 {
		t.Fatalf("expected no failed ops, got %+v", res.Failed)
	}
	if len(calls) != 1 || calls[0].before != "A" || calls[0].after != "B" {
		t.Fatalf("expected one hook call A -> B, got %+v", calls)
	}
}

func TestApplyGrid_AfterUpdateFailureFailsOperation(t *testing.T) {
	store := newFakeStore()
	store.put(&fakeEntity{id: "x1", name: "A", version: 1})
	cfg := cfgFor(store, nil)
	boom := errors.New("invite failed")
	cfg.CRUD.AfterUpdate = func(_ context.Context, _ *iam.Actor, _, _ grid.Row, _ *fakeEntity) error {
		return boom
	}
	uc := usecases.NewApplyGridUseCase(cfg, database.Noop())
	g := sampleGrid()

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{
		SchemaVersion: g.Version(),
		Operations: []grid.Op{
			{Kind: grid.OpUpdate, ID: "x1", BaseVersion: 1, Data: map[grid.ColumnID]any{"name": "B"}},
		},
	})

	test.NoErr(t, err)
	if len(res.Applied) != 0 {
		t.Fatalf("expected no applied ops, got %+v", res.Applied)
	}
	if len(res.Failed) != 1 || !strings.Contains(res.Failed[0].Message, "invite failed") {
		t.Fatalf("expected hook failure, got %+v", res.Failed)
	}
}

func TestApplyGrid_RollsBackOnFailedOp(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	uc := usecases.NewApplyGridUseCase(cfg, database.Noop())
	g := sampleGrid()

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{
		SchemaVersion: g.Version(),
		Operations: []grid.Op{
			{Kind: grid.OpCreate, TempID: "t1", Data: map[grid.ColumnID]any{"name": "Alice"}},
			{Kind: grid.OpUpdate, ID: "missing", BaseVersion: 1, Data: map[grid.ColumnID]any{"name": "X"}},
		},
	})

	test.NoErr(t, err)
	if len(res.Applied) != 0 {
		t.Errorf("expected applied to be cleared on rollback, got %d", len(res.Applied))
	}
	if len(res.Failed) != 1 {
		t.Fatalf("expected one failed op, got %d", len(res.Failed))
	}
	if res.Failed[0].Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %s", res.Failed[0].Code)
	}
}

func TestApplyGrid_VersionMismatchFailsOp(t *testing.T) {
	store := newFakeStore()
	store.put(&fakeEntity{id: "x1", name: "A", version: 5})
	cfg := cfgFor(store, nil)
	uc := usecases.NewApplyGridUseCase(cfg, database.Noop())
	g := sampleGrid()

	res, _ := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{
		SchemaVersion: g.Version(),
		Operations: []grid.Op{
			{Kind: grid.OpUpdate, ID: "x1", BaseVersion: 1, Data: map[grid.ColumnID]any{"name": "B"}},
		},
	})

	if len(res.Failed) != 1 || res.Failed[0].Code != "CONFLICT" {
		t.Errorf("expected CONFLICT failure, got %+v", res.Failed)
	}
}

func TestApplyGrid_TxRunnerErrorPropagates(t *testing.T) {
	store := newFakeStore()
	cfg := cfgFor(store, nil)
	boom := errors.New("tx broken")
	uc := usecases.NewApplyGridUseCase(cfg, &errTxRunner{err: boom})
	g := sampleGrid()

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ApplyGridCommand{
		SchemaVersion: g.Version(),
		Operations:    []grid.Op{{Kind: grid.OpCreate, TempID: "t1", Data: map[grid.ColumnID]any{"name": "A"}}},
	})

	if !errors.Is(err, boom) {
		t.Errorf("expected tx error to propagate, got %v", err)
	}
}

type errTxRunner struct{ err error }

func (e *errTxRunner) Run(_ context.Context, _ func(context.Context) error) error { return e.err }
