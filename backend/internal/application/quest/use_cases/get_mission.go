package usecases

import (
	"context"
	"errors"
	"log/slog"

	"github.com/els/backend/internal/application/quest/runtime"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type GetMissionUseCase struct {
	missions quest.MissionRepository
	dialog   *runtime.Dialog
	logger   *slog.Logger
}

func NewGetMissionUseCase(missions quest.MissionRepository, dialog *runtime.Dialog, logger *slog.Logger) *GetMissionUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetMissionUseCase{missions: missions, dialog: dialog, logger: logger}
}

type GetMissionResult struct {
	Mission     *quest.CustomMission
	ActiveReply *quest.RespondJobStatusResponse
}

func (uc *GetMissionUseCase) Execute(ctx context.Context, actor *iam.Actor, missionID string) (GetMissionResult, error) {
	// 1. Only an authenticated actor reads missions.
	if actor == nil {
		return GetMissionResult{}, shared.ErrUnauthorized
	}
	userID := actor.AccountID().String()

	// 2. Load the personal copy; if missing — fork the shared mission from the catalog.
	mission, err := uc.missions.GetByID(ctx, userID, missionID)
	if errors.Is(err, shared.ErrNotFound) {
		mission, err = uc.fork(ctx, userID, missionID)
	}
	if err != nil {
		return GetMissionResult{}, err
	}

	// 3. Mark stuck image generation as an error so polling does not hang forever.
	if quest.RecoverStaleImageStatuses(mission) {
		err := uc.missions.Update(ctx, userID, missionID, func(fresh *quest.CustomMission) error {
			quest.RecoverStaleImageStatuses(fresh)
			return nil
		})
		if err != nil {
			uc.logger.Warn("quest: save recovered image statuses failed", slog.String("mission", missionID), slog.String("err", err.Error()))
		}
	}

	// 4. Attach the active reply status from the in-memory queue.
	active := uc.dialog.SnapshotByMission(userID, missionID)

	return GetMissionResult{Mission: mission, ActiveReply: active}, nil
}

func (uc *GetMissionUseCase) fork(ctx context.Context, userID, missionID string) (*quest.CustomMission, error) {
	origin, authorID, err := uc.missions.GetOrigin(ctx, missionID)
	if err != nil {
		return nil, err
	}
	if origin.GenerationStatus != quest.GenerationStatusReady {
		return nil, shared.ErrNotFound
	}
	fork := origin.ForkForPlayer()
	if err := uc.missions.Fork(ctx, userID, authorID, fork); err != nil {
		return nil, err
	}
	return fork, nil
}
