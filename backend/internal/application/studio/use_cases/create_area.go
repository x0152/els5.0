package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
	"github.com/els/backend/internal/utils/timex"
)

type CreateAreaCommand struct {
	Title string
	Icon  string
}

type CreateAreaUseCase struct {
	repo  studio.Repository
	clock timex.Clock
}

func NewCreateAreaUseCase(repo studio.Repository, clock timex.Clock) *CreateAreaUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &CreateAreaUseCase{repo: repo, clock: clock}
}

func (uc *CreateAreaUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd CreateAreaCommand) (studio.Area, error) {
	// 1. Build and validate the area.
	area := studio.Area{
		ID:        uuid.NewString(),
		AccountID: actor.AccountID().String(),
		Title:     cmd.Title,
		Icon:      cmd.Icon,
		CreatedAt: uc.clock.Now(),
	}
	if err := area.Validate(); err != nil {
		return studio.Area{}, err
	}

	// 2. Persist.
	if err := uc.repo.InsertArea(ctx, area); err != nil {
		return studio.Area{}, err
	}
	return area, nil
}
