package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type TrainerCheckCommand struct {
	Dialogue string
	Draft    string
	Level    diary.TrainerLevel
}

type TrainerCheckUseCase struct {
	llm LLMClient
}

func NewTrainerCheckUseCase(llm LLMClient) *TrainerCheckUseCase {
	return &TrainerCheckUseCase{llm: llm}
}

func (uc *TrainerCheckUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd TrainerCheckCommand) (diary.TrainerVerdict, error) {
	// 1. Validate the draft and LLM availability.
	if strings.TrimSpace(cmd.Draft) == "" {
		return diary.TrainerVerdict{}, fmt.Errorf("draft is required: %w", shared.ErrValidation)
	}
	if !uc.llm.Available() {
		return diary.TrainerVerdict{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Ask the LLM for a verdict at the requested strictness level.
	system, user := diary.BuildTrainerPrompt(cmd.Dialogue, cmd.Draft, actor.Account().NativeLanguage(), cmd.Level)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return diary.TrainerVerdict{}, err
	}

	// 3. Parse and sanity-check the verdict against the draft.
	return diary.ParseTrainerVerdict(raw, cmd.Draft)
}
