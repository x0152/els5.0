package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/illustration"
)

type Illustrator interface {
	Ensure(ctx context.Context, prompt, aspect string, trigger bool) illustration.Status
}

type EnsureIllustrationCommand struct {
	Prompt  string
	Aspect  string
	Trigger bool
}

type EnsureIllustrationUseCase struct {
	illustrator Illustrator
}

func NewEnsureIllustrationUseCase(illustrator Illustrator) *EnsureIllustrationUseCase {
	return &EnsureIllustrationUseCase{illustrator: illustrator}
}

func (uc *EnsureIllustrationUseCase) Execute(ctx context.Context, cmd EnsureIllustrationCommand) illustration.Status {
	// 1. Trigger generation if requested, or report the current status.
	return uc.illustrator.Ensure(ctx, cmd.Prompt, cmd.Aspect, cmd.Trigger)
}
