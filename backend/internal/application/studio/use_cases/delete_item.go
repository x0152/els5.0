package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
)

type DeleteItemUseCase struct {
	repo studio.Repository
}

func NewDeleteItemUseCase(repo studio.Repository) *DeleteItemUseCase {
	return &DeleteItemUseCase{repo: repo}
}

func (uc *DeleteItemUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) error {
	// 1. Delete the item.
	return uc.repo.Delete(ctx, actor.AccountID().String(), id)
}
