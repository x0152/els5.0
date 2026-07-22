package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/studio"
)

type RegenExampleUseCase struct {
	repo studio.Repository
	llm  LLMClient
}

func NewRegenExampleUseCase(repo studio.Repository, llm LLMClient) *RegenExampleUseCase {
	return &RegenExampleUseCase{repo: repo, llm: llm}
}

func (uc *RegenExampleUseCase) Execute(ctx context.Context, actor *iam.Actor, itemID string) (studio.Item, error) {
	// 1. Load the item and check LLM availability.
	item, err := uc.repo.Get(ctx, actor.AccountID().String(), itemID)
	if err != nil {
		return studio.Item{}, err
	}
	if !uc.llm.Available() {
		return studio.Item{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Generate a fresh example.
	system, user := studio.BuildExamplePrompt(item.Text, item.Example)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return studio.Item{}, err
	}
	example, err := studio.ParseExample(raw)
	if err != nil {
		return studio.Item{}, err
	}
	item.Example = example

	// 3. Persist.
	if err := uc.repo.Update(ctx, item); err != nil {
		return studio.Item{}, err
	}
	return item, nil
}
