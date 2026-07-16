package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/speech"
)

type AssessCommand struct {
	Audio      []byte
	Text       string
	Strictness float64
}

type AssessUseCase struct {
	assessor speech.Assessor
}

func NewAssessUseCase(assessor speech.Assessor) *AssessUseCase {
	return &AssessUseCase{assessor: assessor}
}

func (uc *AssessUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd AssessCommand) (speech.Assessment, error) {
	// 1. Validate the recording and reference text.
	if len(cmd.Audio) == 0 {
		return speech.Assessment{}, fmt.Errorf("audio is required: %w", shared.ErrValidation)
	}
	if strings.TrimSpace(cmd.Text) == "" {
		return speech.Assessment{}, fmt.Errorf("text is required: %w", shared.ErrValidation)
	}
	if cmd.Strictness < speech.MinStrictness || cmd.Strictness > speech.MaxStrictness {
		cmd.Strictness = speech.DefaultStrictness
	}
	// 2. Score the pronunciation via the speech service.
	return uc.assessor.Assess(ctx, cmd.Audio, cmd.Text, cmd.Strictness)
}
