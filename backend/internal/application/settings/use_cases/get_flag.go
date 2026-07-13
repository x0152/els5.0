package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
)

type GetEventProcessingUseCase struct {
	repo settings.FlagRepository
}

func NewGetEventProcessingUseCase(repo settings.FlagRepository) *GetEventProcessingUseCase {
	return &GetEventProcessingUseCase{repo: repo}
}

func (uc *GetEventProcessingUseCase) Execute(ctx context.Context, actor *iam.Actor) (bool, error) {
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return false, err
	}
	return uc.repo.GetFlag(ctx, settings.FlagEventProcessing)
}
