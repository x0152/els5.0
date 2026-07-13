package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type ResetMissionUseCase struct {
	missions quest.MissionRepository
}

func NewResetMissionUseCase(missions quest.MissionRepository) *ResetMissionUseCase {
	return &ResetMissionUseCase{missions: missions}
}

func (uc *ResetMissionUseCase) Execute(ctx context.Context, actor *iam.Actor, missionID string) (*quest.CustomMission, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	userID := actor.AccountID().String()

	mission, err := uc.missions.GetByID(ctx, userID, missionID)
	if err != nil {
		return nil, err
	}

	mission.Reset()

	if err := uc.missions.Save(ctx, userID, mission); err != nil {
		return nil, err
	}
	return mission, nil
}
