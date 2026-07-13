package usecases

import (
	"context"
	"strings"

	"github.com/els/backend/internal/application/quest/runtime"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type StartRespondUseCase struct {
	dialog *runtime.Dialog
}

func NewStartRespondUseCase(dialog *runtime.Dialog) *StartRespondUseCase {
	return &StartRespondUseCase{dialog: dialog}
}

func (uc *StartRespondUseCase) Execute(ctx context.Context, actor *iam.Actor, missionID string, req quest.RespondRequest) (*quest.StartRespondJobResponse, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return nil, shared.ErrValidation
	}
	strict := req.Strict != nil && *req.Strict
	return uc.dialog.Start(ctx, actor.AccountID().String(), missionID, text, strict)
}
