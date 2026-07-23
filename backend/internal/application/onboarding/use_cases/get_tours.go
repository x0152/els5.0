package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/onboarding"
)

type GetToursUseCase struct {
	repo onboarding.Repository
}

func NewGetToursUseCase(repo onboarding.Repository) *GetToursUseCase {
	return &GetToursUseCase{repo: repo}
}

func (uc *GetToursUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]string, error) {
	// 1. Load completed tour ids for the account.
	return uc.repo.Tours(ctx, actor.AccountID().String())
}
