package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type UpdateStatusUseCase struct {
	units vocab.Repository
}

func NewUpdateStatusUseCase(units vocab.Repository) *UpdateStatusUseCase {
	return &UpdateStatusUseCase{units: units}
}

func (uc *UpdateStatusUseCase) Execute(ctx context.Context, actor *iam.Actor, id string, status vocab.Status) (vocab.Unit, error) {
	if actor == nil {
		return vocab.Unit{}, shared.ErrUnauthorized
	}
	if !status.IsValid() {
		return vocab.Unit{}, shared.Validation(fmt.Errorf("status: invalid %q", status))
	}
	return uc.units.UpdateStatus(ctx, actor.AccountID().String(), id, status)
}
