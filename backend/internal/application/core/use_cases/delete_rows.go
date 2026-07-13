package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
)

type DeleteStore interface {
	DeleteRows(ctx context.Context, userID, kind string, ids []string) (int64, error)
}

type DeleteRowsUseCase struct {
	store DeleteStore
}

func NewDeleteRowsUseCase(store DeleteStore) *DeleteRowsUseCase {
	return &DeleteRowsUseCase{store: store}
}

type DeleteRowsCommand struct {
	Kind string
	IDs  []string
}

func (uc *DeleteRowsUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd DeleteRowsCommand) (int64, error) {
	// 1. The words/rules catalog is shared — only a global admin may delete from it.
	if cmd.Kind == "words" || cmd.Kind == "rules" {
		if err := iam.RequireGlobalAdmin(actor); err != nil {
			return 0, err
		}
	}
	if len(cmd.IDs) == 0 {
		return 0, nil
	}

	// 2. Events and raw records are deleted only within the actor's own account.
	return uc.store.DeleteRows(ctx, actor.AccountID().String(), cmd.Kind, cmd.IDs)
}
