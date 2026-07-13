package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/utils/timex"
)

type SaveProgressUseCase struct {
	films films.Repository
	clock timex.Clock
}

func NewSaveProgressUseCase(repo films.Repository, clock timex.Clock) *SaveProgressUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &SaveProgressUseCase{films: repo, clock: clock}
}

func (uc *SaveProgressUseCase) Execute(ctx context.Context, actor *iam.Actor, id string, positionMs int) error {
	// 1. Clamp the position — a negative value means the start of the film.
	if positionMs < 0 {
		positionMs = 0
	}
	// 2. Upsert the owner's watch position.
	return uc.films.SaveProgress(ctx, actor.AccountID().String(), id, positionMs, uc.clock.Now())
}
