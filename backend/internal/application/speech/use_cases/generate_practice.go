package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/speech"
)

type GeneratePracticeCommand struct {
	Topic  string
	Sounds []string
}

type GeneratePracticeUseCase struct {
	llm LLMClient
}

func NewGeneratePracticeUseCase(llm LLMClient) *GeneratePracticeUseCase {
	return &GeneratePracticeUseCase{llm: llm}
}

func (uc *GeneratePracticeUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd GeneratePracticeCommand) ([]string, error) {
	// 1. Ensure the LLM is configured.
	if !uc.llm.Available() {
		return nil, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Generate and parse the practice sentences.
	system, user := speech.BuildPracticePrompt(cmd.Topic, cmd.Sounds)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return nil, err
	}
	return speech.ParsePractice(raw)
}
