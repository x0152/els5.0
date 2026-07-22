package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
	"github.com/els/backend/internal/utils/timex"
)

type PassReviewUseCase struct {
	repo  studio.Repository
	clock timex.Clock
}

func NewPassReviewUseCase(repo studio.Repository, clock timex.Clock) *PassReviewUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &PassReviewUseCase{repo: repo, clock: clock}
}

func (uc *PassReviewUseCase) Execute(ctx context.Context, actor *iam.Actor, itemID string) (studio.Item, error) {
	// 1. Load the item.
	item, err := uc.repo.Get(ctx, actor.AccountID().String(), itemID)
	if err != nil {
		return studio.Item{}, err
	}

	// 2. Advance the review schedule — the entity validates that a review is due.
	if err := item.PassReview(uc.clock.Now()); err != nil {
		return studio.Item{}, err
	}

	// 3. Persist.
	if err := uc.repo.Update(ctx, item); err != nil {
		return studio.Item{}, err
	}
	return item, nil
}
