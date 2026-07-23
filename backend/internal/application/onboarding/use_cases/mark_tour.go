package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/onboarding"
	"github.com/els/backend/internal/utils/timex"
)

type MarkTourUseCase struct {
	repo  onboarding.Repository
	clock timex.Clock
}

func NewMarkTourUseCase(repo onboarding.Repository, clock timex.Clock) *MarkTourUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &MarkTourUseCase{repo: repo, clock: clock}
}

func (uc *MarkTourUseCase) Execute(ctx context.Context, actor *iam.Actor, tourID string) error {
	// 1. Persist the completed tour.
	return uc.repo.SaveTour(ctx, actor.AccountID().String(), tourID, uc.clock.Now())
}
