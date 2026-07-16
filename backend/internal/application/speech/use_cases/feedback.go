package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/speech"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type FeedbackCommand struct {
	Text           string
	Heard          string
	NativeLanguage string
	Issues         []string
}

type FeedbackUseCase struct {
	llm LLMClient
}

func NewFeedbackUseCase(llm LLMClient) *FeedbackUseCase {
	return &FeedbackUseCase{llm: llm}
}

func (uc *FeedbackUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd FeedbackCommand) (speech.Feedback, error) {
	// 1. Validate the transcription pair.
	if strings.TrimSpace(cmd.Text) == "" || strings.TrimSpace(cmd.Heard) == "" {
		return speech.Feedback{}, fmt.Errorf("text and heard transcription are required: %w", shared.ErrValidation)
	}
	if !uc.llm.Available() {
		return speech.Feedback{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}
	// 2. Ask the LLM for coaching advice.
	system, user := speech.BuildFeedbackPrompt(cmd.Text, cmd.Heard, cmd.NativeLanguage, cmd.Issues)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return speech.Feedback{}, err
	}
	// 3. Parse the structured advice.
	return speech.ParseFeedback(raw)
}
