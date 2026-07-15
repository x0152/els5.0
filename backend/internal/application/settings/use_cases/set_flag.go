package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
)

type SetFlagUseCase struct {
	repo settings.FlagRepository
	key  string
}

func NewSetFlagUseCase(repo settings.FlagRepository, key string) *SetFlagUseCase {
	return &SetFlagUseCase{repo: repo, key: key}
}

func (uc *SetFlagUseCase) Execute(ctx context.Context, actor *iam.Actor, enabled bool) error {
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return err
	}
	return uc.repo.SetFlag(ctx, uc.key, enabled)
}
