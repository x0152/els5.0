package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/application/quest/runtime"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type CreateMissionUseCase struct {
	missions  quest.MissionRepository
	generator *runtime.Generator
}

func NewCreateMissionUseCase(missions quest.MissionRepository, generator *runtime.Generator) *CreateMissionUseCase {
	return &CreateMissionUseCase{missions: missions, generator: generator}
}

type CreateMissionResult struct {
	MissionID string
}

func (uc *CreateMissionUseCase) Execute(ctx context.Context, actor *iam.Actor, req quest.CreateMissionRequest) (CreateMissionResult, error) {
	// 1. Only an authenticated actor creates missions.
	if actor == nil {
		return CreateMissionResult{}, shared.ErrUnauthorized
	}
	userID := actor.AccountID().String()

	// 2. Save a stub with status generating — the list shows it immediately.
	mission := &quest.CustomMission{
		ID:               uuid.NewString(),
		UserPrompt:       req.Prompt,
		Genre:            req.Genre,
		Language:         req.Language,
		PracticeGoals:    req.PracticeGoals,
		CreatedAt:        time.Now().Format(time.RFC3339),
		GenerationStatus: quest.GenerationStatusGenerating,
		GenerationStep:   quest.GenerationStepCreating,
	}
	if err := uc.missions.Insert(ctx, userID, mission); err != nil {
		return CreateMissionResult{}, err
	}

	// 3. Generation runs in the background (goroutine); status is read via polling.
	uc.generator.Enqueue(userID, mission.ID, req)

	return CreateMissionResult{MissionID: mission.ID}, nil
}
