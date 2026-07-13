package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type GetFilmUseCase struct {
	films films.Repository
}

func NewGetFilmUseCase(repo films.Repository) *GetFilmUseCase {
	return &GetFilmUseCase{films: repo}
}

func (uc *GetFilmUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) (films.Film, int, error) {
	if actor == nil {
		return films.Film{}, 0, shared.ErrUnauthorized
	}
	// 1. Load the film.
	film, err := uc.films.Get(ctx, id)
	if err != nil {
		return films.Film{}, 0, err
	}
	// 2. Load the actor's watch position for it.
	progress, err := uc.films.ListProgress(ctx, actor.AccountID().String())
	if err != nil {
		return films.Film{}, 0, err
	}
	return film, progress[film.ID], nil
}
