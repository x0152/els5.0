package usecases

import (
	"context"

	"github.com/els/backend/internal/application/quest/runtime"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type DeleteMissionUseCase struct {
	missions quest.MissionRepository
	images   *runtime.Images
}

func NewDeleteMissionUseCase(missions quest.MissionRepository, images *runtime.Images) *DeleteMissionUseCase {
	return &DeleteMissionUseCase{missions: missions, images: images}
}

func (uc *DeleteMissionUseCase) Execute(ctx context.Context, actor *iam.Actor, missionID string) error {
	// 1. The catalog is shared: any authenticated user may delete a mission.
	if actor == nil {
		return shared.ErrUnauthorized
	}

	// 2. Collect all players' copies to clean up their images.
	copies, err := uc.missions.GetAllByID(ctx, missionID)
	if err != nil {
		return err
	}
	if len(copies) == 0 {
		return shared.ErrNotFound
	}

	// 3. Delete the original — forks are removed by cascade.
	if err := uc.missions.Delete(ctx, missionID); err != nil {
		return err
	}

	if uc.images != nil {
		for _, m := range copies {
			uc.images.DeleteMissionImages(ctx, m)
		}
	}
	return nil
}
