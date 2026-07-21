package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/reading"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type WordSource interface {
	List(ctx context.Context, accountID string, filter vocab.ListFilter) ([]vocab.Unit, int, error)
}

type GenerateTextCommand struct {
	Topic    string
	UseVocab bool
	Level    reading.Level
	Length   reading.Length
}

type GenerateTextUseCase struct {
	llm   LLMClient
	words WordSource
}

func NewGenerateTextUseCase(llm LLMClient, words WordSource) *GenerateTextUseCase {
	return &GenerateTextUseCase{llm: llm, words: words}
}

func (uc *GenerateTextUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd GenerateTextCommand) (reading.Text, error) {
	// 1. Ensure the LLM is configured.
	if !uc.llm.Available() {
		return reading.Text{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Optionally pick up words the learner is studying.
	var words []string
	if cmd.UseVocab && uc.words != nil {
		units, _, err := uc.words.List(ctx, actor.AccountID().String(), vocab.ListFilter{Status: vocab.StatusLearning, Limit: 10})
		if err == nil {
			for _, u := range units {
				words = append(words, u.Text)
			}
		}
	}

	// 3. Generate and parse the text.
	system, user := reading.BuildTextPrompt(cmd.Topic, words, cmd.Level, cmd.Length)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return reading.Text{}, err
	}
	return reading.ParseText(raw)
}
