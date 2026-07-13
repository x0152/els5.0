package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type DeleteUnitUseCase struct {
	units vocab.Repository
}

func NewDeleteUnitUseCase(units vocab.Repository) *DeleteUnitUseCase {
	return &DeleteUnitUseCase{units: units}
}

func (uc *DeleteUnitUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) error {
	if actor == nil {
		return shared.ErrUnauthorized
	}
	return uc.units.Delete(ctx, actor.AccountID().String(), id)
}
