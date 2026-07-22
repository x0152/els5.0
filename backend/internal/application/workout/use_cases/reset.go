package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/workout"
)

type ResetUseCase struct {
	repo workout.Repository
}

func NewResetUseCase(repo workout.Repository) *ResetUseCase {
	return &ResetUseCase{repo: repo}
}

func (uc *ResetUseCase) Execute(ctx context.Context, actor *iam.Actor) error {
	// 1. Delete all workout data of the account: lessons, item stats and film positions.
	return uc.repo.DeleteAccountData(ctx, actor.AccountID().String())
}
