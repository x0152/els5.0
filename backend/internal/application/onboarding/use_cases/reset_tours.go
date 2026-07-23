package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/onboarding"
)

type ResetToursUseCase struct {
	repo onboarding.Repository
}

func NewResetToursUseCase(repo onboarding.Repository) *ResetToursUseCase {
	return &ResetToursUseCase{repo: repo}
}

func (uc *ResetToursUseCase) Execute(ctx context.Context, actor *iam.Actor) error {
	// 1. Delete all completed tours so onboarding shows again.
	return uc.repo.DeleteTours(ctx, actor.AccountID().String())
}
