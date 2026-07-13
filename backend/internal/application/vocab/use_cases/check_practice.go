package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type CheckPracticeUseCase struct {
	llm LLMClient
}

func NewCheckPracticeUseCase(llm LLMClient) *CheckPracticeUseCase {
	return &CheckPracticeUseCase{llm: llm}
}

func (uc *CheckPracticeUseCase) Execute(ctx context.Context, actor *iam.Actor, instruction, answer string) (vocab.PracticeCheckResult, error) {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return vocab.PracticeCheckResult{}, shared.Validation(fmt.Errorf("answer: must not be empty"))
	}
	if !uc.llm.Available() {
		return vocab.PracticeCheckResult{}, shared.ErrUnavailable
	}

	// 1. Send the free-form answer to the LLM for checking.
	system, user := vocab.BuildPracticeCheckPrompt(strings.TrimSpace(instruction), answer)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return vocab.PracticeCheckResult{}, err
	}
	return vocab.ParsePracticeCheckResult(raw)
}
