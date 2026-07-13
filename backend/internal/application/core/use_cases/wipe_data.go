package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
)

type WipeStore interface {
	WipeUser(ctx context.Context, userID string) error
}

type WipeDataUseCase struct {
	store WipeStore
}

func NewWipeDataUseCase(store WipeStore) *WipeDataUseCase {
	return &WipeDataUseCase{store: store}
}

func (uc *WipeDataUseCase) Execute(ctx context.Context, actor *iam.Actor) error {
	// 1. Each actor clears only their own events; the shared catalog is untouched.
	return uc.store.WipeUser(ctx, actor.AccountID().String())
}
