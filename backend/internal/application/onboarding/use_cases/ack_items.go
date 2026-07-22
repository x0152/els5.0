package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/onboarding"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/timex"
)

type AckItemsUseCase struct {
	repo  onboarding.Repository
	clock timex.Clock
}

func NewAckItemsUseCase(repo onboarding.Repository, clock timex.Clock) *AckItemsUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &AckItemsUseCase{repo: repo, clock: clock}
}

func (uc *AckItemsUseCase) Execute(ctx context.Context, actor *iam.Actor, itemIDs []string) error {
	// 1. Reject unknown item ids.
	for _, id := range itemIDs {
		if !onboarding.ValidItemID(id) {
			return fmt.Errorf("%w: unknown item id %q", shared.ErrValidation, id)
		}
	}

	// 2. Persist acknowledgements.
	if len(itemIDs) == 0 {
		return nil
	}
	return uc.repo.SaveAcks(ctx, actor.AccountID().String(), itemIDs, uc.clock.Now())
}
