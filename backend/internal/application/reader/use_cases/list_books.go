package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type ListBooksUseCase struct {
	books reader.Repository
}

func NewListBooksUseCase(repo reader.Repository) *ListBooksUseCase {
	return &ListBooksUseCase{books: repo}
}

func (uc *ListBooksUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]reader.Book, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	return uc.books.List(ctx, actor.AccountID().String())
}
