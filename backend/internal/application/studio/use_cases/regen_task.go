package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/studio"
)

type RegenTaskUseCase struct {
	repo studio.Repository
	llm  LLMClient
}

func NewRegenTaskUseCase(repo studio.Repository, llm LLMClient) *RegenTaskUseCase {
	return &RegenTaskUseCase{repo: repo, llm: llm}
}

func (uc *RegenTaskUseCase) Execute(ctx context.Context, actor *iam.Actor, itemID string) (studio.Item, error) {
	// 1. Load the item and check LLM availability.
	item, err := uc.repo.Get(ctx, actor.AccountID().String(), itemID)
	if err != nil {
		return studio.Item{}, err
	}
	if !uc.llm.Available() {
		return studio.Item{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Generate a fresh "use it" mini-situation.
	system, user := studio.BuildTaskPrompt(item.Text, item.Task)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return studio.Item{}, err
	}
	task, err := studio.ParseTask(raw)
	if err != nil {
		return studio.Item{}, err
	}
	item.Task = task

	// 3. Persist.
	if err := uc.repo.Update(ctx, item); err != nil {
		return studio.Item{}, err
	}
	return item, nil
}
