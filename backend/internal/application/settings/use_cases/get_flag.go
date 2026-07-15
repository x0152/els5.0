package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
)

type GetFlagUseCase struct {
	repo settings.FlagRepository
	key  string
}

func NewGetFlagUseCase(repo settings.FlagRepository, key string) *GetFlagUseCase {
	return &GetFlagUseCase{repo: repo, key: key}
}

func (uc *GetFlagUseCase) Execute(ctx context.Context, actor *iam.Actor) (bool, error) {
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return false, err
	}
	return uc.repo.GetFlag(ctx, uc.key)
}
