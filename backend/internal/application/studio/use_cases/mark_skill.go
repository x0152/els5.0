package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
	"github.com/els/backend/internal/utils/timex"
)

type MarkSkillCommand struct {
	ItemID string
	Skill  string
}

type MarkSkillUseCase struct {
	repo  studio.Repository
	clock timex.Clock
}

func NewMarkSkillUseCase(repo studio.Repository, clock timex.Clock) *MarkSkillUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &MarkSkillUseCase{repo: repo, clock: clock}
}

func (uc *MarkSkillUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd MarkSkillCommand) (studio.Item, error) {
	// 1. Load the item.
	item, err := uc.repo.Get(ctx, actor.AccountID().String(), cmd.ItemID)
	if err != nil {
		return studio.Item{}, err
	}

	// 2. Mark the skill — the entity validates the skill name.
	if err := item.MarkSkill(cmd.Skill); err != nil {
		return studio.Item{}, err
	}

	// 3. Mastering all skills schedules the first review.
	item.ScheduleReviewIfDone(uc.clock.Now())

	// 4. Persist.
	if err := uc.repo.Update(ctx, item); err != nil {
		return studio.Item{}, err
	}
	return item, nil
}
