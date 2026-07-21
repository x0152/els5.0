package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/writing"
)

type GenerateSituationCommand struct {
	Topic string
}

type GenerateSituationUseCase struct {
	llm LLMClient
}

func NewGenerateSituationUseCase(llm LLMClient) *GenerateSituationUseCase {
	return &GenerateSituationUseCase{llm: llm}
}

func (uc *GenerateSituationUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd GenerateSituationCommand) (writing.Situation, error) {
	// 1. Ensure the LLM is configured.
	if !uc.llm.Available() {
		return writing.Situation{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Generate and parse the situation.
	system, user := writing.BuildSituationPrompt(cmd.Topic)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return writing.Situation{}, err
	}
	return writing.ParseSituation(raw)
}
