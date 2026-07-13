package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
)

type MeUseCase struct{}

func NewMeUseCase() *MeUseCase { return &MeUseCase{} }

type MeResult struct {
	Actor *iam.Actor
}

func (uc *MeUseCase) Execute(_ context.Context, actor *iam.Actor) (MeResult, error) {
	return MeResult{Actor: actor}, nil
}
