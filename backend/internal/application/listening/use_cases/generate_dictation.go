package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/listening"
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

type GenerateDictationCommand struct {
	Topic    string
	UseVocab bool
	Level    listening.Level
	Count    int
}

type GenerateDictationUseCase struct {
	llm   LLMClient
	words WordSource
}

func NewGenerateDictationUseCase(llm LLMClient, words WordSource) *GenerateDictationUseCase {
	return &GenerateDictationUseCase{llm: llm, words: words}
}

func (uc *GenerateDictationUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd GenerateDictationCommand) (listening.Dictation, error) {
	// 1. Ensure the LLM is configured.
	if !uc.llm.Available() {
		return listening.Dictation{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Optionally pick up words the learner is studying.
	var words []string
	if cmd.UseVocab && uc.words != nil {
		units, _, err := uc.words.List(ctx, actor.AccountID().String(), vocab.ListFilter{Status: vocab.StatusLearning, Limit: 8})
		if err == nil {
			for _, u := range units {
				words = append(words, u.Text)
			}
		}
	}

	// 3. Generate and parse the dictation.
	system, user := listening.BuildDictationPrompt(cmd.Topic, words, cmd.Level, cmd.Count)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return listening.Dictation{}, err
	}
	return listening.ParseDictation(raw)
}
