package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
)

type SetEventProcessingUseCase struct {
	repo settings.FlagRepository
}

func NewSetEventProcessingUseCase(repo settings.FlagRepository) *SetEventProcessingUseCase {
	return &SetEventProcessingUseCase{repo: repo}
}

func (uc *SetEventProcessingUseCase) Execute(ctx context.Context, actor *iam.Actor, enabled bool) error {
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return err
	}
	return uc.repo.SetFlag(ctx, settings.FlagEventProcessing, enabled)
}
