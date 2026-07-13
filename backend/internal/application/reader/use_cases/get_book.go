package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type GetBookUseCase struct {
	books reader.Repository
}

func NewGetBookUseCase(repo reader.Repository) *GetBookUseCase {
	return &GetBookUseCase{books: repo}
}

func (uc *GetBookUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) (reader.Book, error) {
	if actor == nil {
		return reader.Book{}, shared.ErrUnauthorized
	}
	return uc.books.Get(ctx, actor.AccountID().String(), id)
}
