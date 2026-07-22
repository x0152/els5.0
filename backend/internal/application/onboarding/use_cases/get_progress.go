package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/onboarding"
	"github.com/els/backend/internal/utils/timex"
)

type GetProgressUseCase struct {
	repo  onboarding.Repository
	clock timex.Clock
}

func NewGetProgressUseCase(repo onboarding.Repository, clock timex.Clock) *GetProgressUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &GetProgressUseCase{repo: repo, clock: clock}
}

func (uc *GetProgressUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]onboarding.Status, error) {
	accountID := actor.AccountID().String()

	// 1. Load stored high-water marks.
	stored, err := uc.repo.Watermarks(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// 2. Count live activity in the source tables.
	live, err := uc.repo.LiveCounts(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// 3. Merge: progress only grows, deletions never roll it back.
	merged := onboarding.Merge(stored, live)

	// 4. Persist metrics that increased.
	if inc := onboarding.Increased(stored, merged); len(inc) > 0 {
		if err := uc.repo.SaveWatermarks(ctx, accountID, inc, uc.clock.Now()); err != nil {
			return nil, err
		}
	}

	// 5. Load acknowledged items and build statuses.
	acked, err := uc.repo.Acks(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return onboarding.Statuses(merged, acked), nil
}
