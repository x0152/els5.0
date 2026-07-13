package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/els/backend/internal/application/grid_engine/gridspec"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/database"
)

type ApplyGridUseCase[E any] struct {
	cfg gridspec.Config[E]
	tx  database.TxRunner
}

func NewApplyGridUseCase[E any](cfg gridspec.Config[E], tx database.TxRunner) *ApplyGridUseCase[E] {
	if tx == nil {
		tx = database.Noop()
	}
	return &ApplyGridUseCase[E]{cfg: cfg, tx: tx}
}

type ApplyGridCommand struct {
	SchemaVersion string
	Operations    []grid.Op
}

var errBulkRollback = errors.New("grid_engine: bulk rollback")

func (uc *ApplyGridUseCase[E]) Execute(ctx context.Context, actor *iam.Actor, cmd ApplyGridCommand) (ApplyResult, error) {
	// 1. Check the actor's access to this specific grid.
	if uc.cfg.Authorize != nil {
		if err := uc.cfg.Authorize(actor); err != nil {
			return ApplyResult{}, err
		}
	}

	// 2. Compare the client's schema_version with the current one; mismatch — CONFLICT.
	g := uc.cfg.Grid(actor)
	if err := grid.EnforceSchemaVersion(cmd.SchemaVersion, g.Version()); err != nil {
		return ApplyResult{}, err
	}

	// 3. Prepare the result accumulator.
	res := ApplyResult{
		SchemaVersion: g.Version(),
		Applied:       make([]grid.OpResult, 0, len(cmd.Operations)),
		Failed:        make([]grid.OpError, 0),
	}

	// 4. Run operations in one transaction; any error → roll back the whole batch.
	txErr := uc.tx.Run(ctx, func(ctx context.Context) error {
		for i, op := range cmd.Operations {
			out, opErr := uc.runOp(ctx, actor, g, i, op)
			if opErr != nil {
				res.Failed = append(res.Failed, *opErr)
				continue
			}
			res.Applied = append(res.Applied, out)
		}
		if len(res.Failed) > 0 {
			return errBulkRollback
		}
		return nil
	})

	// 5. On rollback return only failed (clear applied).
	if errors.Is(txErr, errBulkRollback) {
		res.Applied = res.Applied[:0]
		return res, nil
	}
	if txErr != nil {
		return ApplyResult{}, txErr
	}
	return res, nil
}

func (uc *ApplyGridUseCase[E]) runOp(ctx context.Context, actor *iam.Actor, g grid.Grid[E], i int, op grid.Op) (grid.OpResult, *grid.OpError) {
	if err := grid.ValidateOp(op, g.Schema()); err != nil {
		return grid.OpResult{}, opError(i, op, err)
	}
	switch op.Kind {
	case grid.OpCreate:
		return uc.runCreate(ctx, actor, g, i, op)
	case grid.OpUpdate:
		return uc.runUpdate(ctx, actor, g, i, op)
	case grid.OpDelete:
		return uc.runDelete(ctx, actor, i, op)
	}
	return grid.OpResult{}, opError(i, op, shared.Validation(fmt.Errorf("op.kind: invalid %q", op.Kind)))
}

func (uc *ApplyGridUseCase[E]) runCreate(ctx context.Context, actor *iam.Actor, g grid.Grid[E], i int, op grid.Op) (grid.OpResult, *grid.OpError) {
	if uc.cfg.CRUD.Create == nil {
		return grid.OpResult{}, opError(i, op, fmt.Errorf("%w: create is not supported", shared.ErrForbidden))
	}
	entity, err := uc.cfg.CRUD.Create(ctx, actor, op.Data)
	if err != nil {
		return grid.OpResult{}, opError(i, op, err)
	}
	row := g.RowOf(entity)
	return grid.OpResult{
		Index:       i,
		Kind:        grid.OpCreate,
		TempID:      op.TempID,
		ID:          row.ID,
		BaseVersion: row.BaseVersion,
	}, nil
}

func (uc *ApplyGridUseCase[E]) runUpdate(ctx context.Context, actor *iam.Actor, g grid.Grid[E], i int, op grid.Op) (grid.OpResult, *grid.OpError) {
	if uc.cfg.CRUD.GetByID == nil || uc.cfg.CRUD.Update == nil || uc.cfg.CRUD.Version == nil {
		return grid.OpResult{}, opError(i, op, fmt.Errorf("%w: update is not supported", shared.ErrForbidden))
	}
	entity, err := uc.cfg.CRUD.GetByID(ctx, actor, op.ID)
	if err != nil {
		return grid.OpResult{}, opError(i, op, err)
	}
	if err := grid.EnforceRowVersion(op.BaseVersion, uc.cfg.CRUD.Version(entity)); err != nil {
		return grid.OpResult{}, opError(i, op, err)
	}
	before := g.RowOf(entity)
	if err := g.ApplyPatch(entity, op.Data); err != nil {
		return grid.OpResult{}, opError(i, op, err)
	}
	if err := uc.cfg.CRUD.Update(ctx, entity); err != nil {
		return grid.OpResult{}, opError(i, op, err)
	}
	row := g.RowOf(entity)
	if uc.cfg.CRUD.AfterUpdate != nil {
		if err := uc.cfg.CRUD.AfterUpdate(ctx, actor, before, row, entity); err != nil {
			return grid.OpResult{}, opError(i, op, err)
		}
	}
	return grid.OpResult{
		Index:       i,
		Kind:        grid.OpUpdate,
		ID:          row.ID,
		BaseVersion: row.BaseVersion,
	}, nil
}

func (uc *ApplyGridUseCase[E]) runDelete(ctx context.Context, actor *iam.Actor, i int, op grid.Op) (grid.OpResult, *grid.OpError) {
	if uc.cfg.CRUD.Delete == nil {
		return grid.OpResult{}, opError(i, op, fmt.Errorf("%w: delete is not supported", shared.ErrForbidden))
	}
	if uc.cfg.CRUD.GetByID != nil && uc.cfg.CRUD.Version != nil {
		entity, err := uc.cfg.CRUD.GetByID(ctx, actor, op.ID)
		if err != nil {
			return grid.OpResult{}, opError(i, op, err)
		}
		if err := grid.EnforceRowVersion(op.BaseVersion, uc.cfg.CRUD.Version(entity)); err != nil {
			return grid.OpResult{}, opError(i, op, err)
		}
	}
	if err := uc.cfg.CRUD.Delete(ctx, actor, op.ID); err != nil {
		return grid.OpResult{}, opError(i, op, err)
	}
	return grid.OpResult{Index: i, Kind: grid.OpDelete, ID: op.ID}, nil
}

func opError(i int, op grid.Op, err error) *grid.OpError {
	return &grid.OpError{
		Index:   i,
		TempID:  op.TempID,
		ID:      op.ID,
		Code:    codeOf(err),
		Message: err.Error(),
	}
}

func codeOf(err error) string {
	switch {
	case errors.Is(err, shared.ErrValidation):
		return "VALIDATION_ERROR"
	case errors.Is(err, shared.ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, shared.ErrConflict):
		return "CONFLICT"
	case errors.Is(err, shared.ErrForbidden):
		return "FORBIDDEN"
	case errors.Is(err, shared.ErrUnauthorized):
		return "UNAUTHORIZED"
	case errors.Is(err, shared.ErrUnavailable):
		return "SERVICE_UNAVAILABLE"
	}
	return "INTERNAL_ERROR"
}
