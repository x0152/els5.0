package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type SaveProgressUseCase struct {
	books reader.Repository
}

func NewSaveProgressUseCase(repo reader.Repository) *SaveProgressUseCase {
	return &SaveProgressUseCase{books: repo}
}

func (uc *SaveProgressUseCase) Execute(ctx context.Context, actor *iam.Actor, id string, position int) error {
	if actor == nil {
		return shared.ErrUnauthorized
	}
	if position < 0 {
		position = 0
	}
	return uc.books.SavePosition(ctx, actor.AccountID().String(), id, position)
}
