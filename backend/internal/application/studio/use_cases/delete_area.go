package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
)

type DeleteAreaUseCase struct {
	repo studio.Repository
}

func NewDeleteAreaUseCase(repo studio.Repository) *DeleteAreaUseCase {
	return &DeleteAreaUseCase{repo: repo}
}

func (uc *DeleteAreaUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) error {
	// 1. Delete the area (items cascade).
	return uc.repo.DeleteArea(ctx, actor.AccountID().String(), id)
}
