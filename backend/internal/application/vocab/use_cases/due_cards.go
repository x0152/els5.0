package usecases

import (
	"context"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/vocab"
	"github.com/els/backend/internal/utils/timex"
)

type DueCardsUseCase struct {
	units vocab.Repository
	clock timex.Clock
}

func NewDueCardsUseCase(units vocab.Repository, clock timex.Clock) *DueCardsUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &DueCardsUseCase{units: units, clock: clock}
}

func (uc *DueCardsUseCase) Execute(ctx context.Context, actor *iam.Actor) (int, error) {
	// 1. Count words that were not answered yet today: they can still advance their streak.
	now := uc.clock.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return uc.units.CountDue(ctx, actor.AccountID().String(), startOfDay)
}
