package usecases

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
	"github.com/els/backend/internal/utils/timex"
)

type CaptureItemCommand struct {
	Text      string
	AreaTitle string
	Icon      string
}

type CaptureItemUseCase struct {
	repo  studio.Repository
	llm   LLMClient
	clock timex.Clock
}

func NewCaptureItemUseCase(repo studio.Repository, llm LLMClient, clock timex.Clock) *CaptureItemUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &CaptureItemUseCase{repo: repo, llm: llm, clock: clock}
}

func (uc *CaptureItemUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd CaptureItemCommand) (studio.Item, error) {
	accountID := actor.AccountID().String()

	// 1. Find the target area by title, create it on first capture.
	areas, err := uc.repo.ListAreas(ctx, accountID)
	if err != nil {
		return studio.Item{}, err
	}
	var areaID string
	for _, a := range areas {
		if strings.EqualFold(a.Title, cmd.AreaTitle) {
			areaID = a.ID
			break
		}
	}
	if areaID == "" {
		area := studio.Area{
			ID:        uuid.NewString(),
			AccountID: accountID,
			Title:     cmd.AreaTitle,
			Icon:      cmd.Icon,
			CreatedAt: uc.clock.Now(),
		}
		if err := area.Validate(); err != nil {
			return studio.Item{}, err
		}
		if err := uc.repo.InsertArea(ctx, area); err != nil {
			return studio.Item{}, err
		}
		areaID = area.ID
	}

	// 2. Build, enrich and validate the item.
	item, err := buildEnrichedItem(ctx, uc.llm, actor, areaID, cmd.Text, uc.clock.Now())
	if err != nil {
		return studio.Item{}, err
	}

	// 3. Persist.
	if err := uc.repo.Insert(ctx, item); err != nil {
		return studio.Item{}, err
	}
	return item, nil
}
