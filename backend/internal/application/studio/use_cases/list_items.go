package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
)

type ListItemsUseCase struct {
	repo studio.Repository
}

func NewListItemsUseCase(repo studio.Repository) *ListItemsUseCase {
	return &ListItemsUseCase{repo: repo}
}

func (uc *ListItemsUseCase) Execute(ctx context.Context, actor *iam.Actor, areaID string) ([]studio.Item, error) {
	// 1. List the area's items.
	return uc.repo.ListItems(ctx, actor.AccountID().String(), areaID)
}
