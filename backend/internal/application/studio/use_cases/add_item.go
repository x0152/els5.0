package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/studio"
	"github.com/els/backend/internal/utils/timex"
)

type AddItemCommand struct {
	AreaID string
	Text   string
}

type AddItemUseCase struct {
	repo  studio.Repository
	llm   LLMClient
	clock timex.Clock
}

func NewAddItemUseCase(repo studio.Repository, llm LLMClient, clock timex.Clock) *AddItemUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &AddItemUseCase{repo: repo, llm: llm, clock: clock}
}

func (uc *AddItemUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd AddItemCommand) (studio.Item, error) {
	// 1. Build, enrich and validate the item.
	item, err := buildEnrichedItem(ctx, uc.llm, actor, cmd.AreaID, cmd.Text, uc.clock.Now())
	if err != nil {
		return studio.Item{}, err
	}

	// 2. Persist.
	if err := uc.repo.Insert(ctx, item); err != nil {
		return studio.Item{}, err
	}
	return item, nil
}

func buildEnrichedItem(ctx context.Context, llm LLMClient, actor *iam.Actor, areaID, text string, now time.Time) (studio.Item, error) {
	item := studio.Item{
		ID:        uuid.NewString(),
		AreaID:    areaID,
		AccountID: actor.AccountID().String(),
		Text:      text,
		CreatedAt: now,
	}
	if err := item.Validate(); err != nil {
		return studio.Item{}, err
	}
	if !llm.Available() {
		return studio.Item{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	system, user := studio.BuildEnrichPrompt(text, actor.Account().NativeLanguage())
	raw, err := llm.Chat(ctx, system, user)
	if err != nil {
		return studio.Item{}, err
	}
	enrichment, err := studio.ParseEnrichment(raw)
	if err != nil {
		return studio.Item{}, err
	}
	item.Transcription = enrichment.Transcription
	item.Translation = enrichment.Translation
	item.Explanation = enrichment.Explanation
	item.ExplanationNative = enrichment.ExplanationNative
	item.Example = enrichment.Example
	return item, nil
}
