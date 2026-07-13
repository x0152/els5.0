package usecases

import (
	"context"

	"github.com/els/backend/internal/application/quest/runtime"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type RegenerateImagesCommand struct {
	Kind string
	Key  string
}

type RegenerateImagesUseCase struct {
	missions quest.MissionRepository
	images   *runtime.Images
}

func NewRegenerateImagesUseCase(missions quest.MissionRepository, images *runtime.Images) *RegenerateImagesUseCase {
	return &RegenerateImagesUseCase{missions: missions, images: images}
}

func (uc *RegenerateImagesUseCase) Execute(ctx context.Context, actor *iam.Actor, missionID string, cmd RegenerateImagesCommand) (*quest.CustomMission, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	userID := actor.AccountID().String()

	mission, err := uc.missions.GetByID(ctx, userID, missionID)
	if err != nil {
		return nil, err
	}

	if uc.images != nil {
		if cmd.Kind == "" {
			uc.images.RegenerateFailed(userID, mission)
		} else {
			uc.images.RegenerateOne(userID, mission, cmd.Kind, cmd.Key)
		}
		// Statuses are marked "generating" synchronously, so re-read to return them.
		if updated, err := uc.missions.GetByID(ctx, userID, missionID); err == nil {
			mission = updated
		}
	}
	return mission, nil
}
