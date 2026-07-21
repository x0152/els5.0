package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/writing"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type TrainerCheckCommand struct {
	Dialogue string
	Draft    string
	Level    writing.TrainerLevel
}

type TrainerCheckUseCase struct {
	llm LLMClient
}

func NewTrainerCheckUseCase(llm LLMClient) *TrainerCheckUseCase {
	return &TrainerCheckUseCase{llm: llm}
}

func (uc *TrainerCheckUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd TrainerCheckCommand) (writing.TrainerVerdict, error) {
	// 1. Validate the draft and LLM availability.
	if strings.TrimSpace(cmd.Draft) == "" {
		return writing.TrainerVerdict{}, fmt.Errorf("draft is required: %w", shared.ErrValidation)
	}
	if !uc.llm.Available() {
		return writing.TrainerVerdict{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Ask the LLM for a verdict at the requested strictness level.
	system, user := writing.BuildTrainerPrompt(cmd.Dialogue, cmd.Draft, actor.Account().NativeLanguage(), cmd.Level)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return writing.TrainerVerdict{}, err
	}

	// 3. Parse and sanity-check the verdict against the draft.
	return writing.ParseTrainerVerdict(raw, cmd.Draft)
}
