package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type ListMissionsUseCase struct {
	missions quest.MissionRepository
}

func NewListMissionsUseCase(missions quest.MissionRepository) *ListMissionsUseCase {
	return &ListMissionsUseCase{missions: missions}
}

func (uc *ListMissionsUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]quest.MissionCatalogItem, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	return uc.missions.List(ctx, actor.AccountID().String())
}
