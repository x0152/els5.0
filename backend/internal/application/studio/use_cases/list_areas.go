package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
	"github.com/els/backend/internal/utils/timex"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type ListAreasUseCase struct {
	repo  studio.Repository
	clock timex.Clock
}

func NewListAreasUseCase(repo studio.Repository, clock timex.Clock) *ListAreasUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &ListAreasUseCase{repo: repo, clock: clock}
}

func (uc *ListAreasUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]studio.AreaStats, error) {
	// 1. List the account's areas with item stats.
	accountID := actor.AccountID().String()
	areas, err := uc.repo.ListAreas(ctx, accountID)
	if err != nil || len(areas) > 0 {
		return areas, err
	}

	// 2. First visit: seed starter areas with ready-made phrases.
	now := uc.clock.Now()
	for _, sa := range studio.Seed {
		area := studio.Area{
			ID:        uuid.NewString(),
			AccountID: accountID,
			Title:     sa.Title,
			Icon:      sa.Icon,
			CreatedAt: now,
		}
		if err := uc.repo.InsertArea(ctx, area); err != nil {
			return nil, err
		}
		for _, si := range sa.Items {
			now = now.Add(time.Second)
			item := studio.Item{
				ID:                uuid.NewString(),
				AreaID:            area.ID,
				AccountID:         accountID,
				Text:              si.Text,
				Transcription:     si.Transcription,
				Translation:       si.Translation,
				Explanation:       si.Explanation,
				ExplanationNative: si.ExplanationNative,
				Example:           si.Example,
				CreatedAt:         now,
			}
			if err := uc.repo.Insert(ctx, item); err != nil {
				return nil, err
			}
		}
		now = now.Add(time.Second)
	}
	return uc.repo.ListAreas(ctx, accountID)
}
