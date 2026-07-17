package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/domain/iam"
)

type ResetHistoryUseCase struct {
	repo diary.Repository
}

func NewResetHistoryUseCase(repo diary.Repository) *ResetHistoryUseCase {
	return &ResetHistoryUseCase{repo: repo}
}

func (uc *ResetHistoryUseCase) Execute(ctx context.Context, actor *iam.Actor) error {
	// 1. Delete all entries owned by the account.
	return uc.repo.DeleteAll(ctx, actor.AccountID().String())
}
