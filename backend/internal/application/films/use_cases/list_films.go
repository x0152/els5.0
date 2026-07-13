package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type ListFilmsUseCase struct {
	films films.Repository
}

func NewListFilmsUseCase(repo films.Repository) *ListFilmsUseCase {
	return &ListFilmsUseCase{films: repo}
}

func (uc *ListFilmsUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]films.Film, map[string]int, error) {
	if actor == nil {
		return nil, nil, shared.ErrUnauthorized
	}
	// 1. Load the catalog.
	list, err := uc.films.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	// 2. Load the actor's watch positions.
	progress, err := uc.films.ListProgress(ctx, actor.AccountID().String())
	if err != nil {
		return nil, nil, err
	}
	return list, progress, nil
}
