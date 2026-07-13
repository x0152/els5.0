package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/templateapp"
)

type EchoUseCase struct{}

func NewEchoUseCase() *EchoUseCase {
	return &EchoUseCase{}
}

type EchoCommand struct {
	Message string
}

type EchoResult struct {
	Message string
}

func (uc *EchoUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd EchoCommand) (EchoResult, error) {
	_ = ctx
	// 1. Actor is required.
	if actor == nil {
		return EchoResult{}, shared.ErrForbidden
	}
	// 2. Domain normalization and length limit.
	m := templateapp.NormalizeEchoMessage(cmd.Message)
	if m == "" {
		return EchoResult{}, fmt.Errorf("%w: message must not be empty", shared.ErrValidation)
	}
	if len(m) > templateapp.MaxEchoMessageLen {
		return EchoResult{}, fmt.Errorf("%w: message too long", shared.ErrValidation)
	}
	return EchoResult{Message: m}, nil
}
