package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type CheckEntryUseCase struct {
	llm LLMClient
}

func NewCheckEntryUseCase(llm LLMClient) *CheckEntryUseCase {
	return &CheckEntryUseCase{llm: llm}
}

// Execute runs the lenient quest grammar check over a diary draft: it ignores
// typos, punctuation and style, flagging only real grammar mistakes.
func (uc *CheckEntryUseCase) Execute(ctx context.Context, actor *iam.Actor, text string) (quest.GrammarCheck, error) {
	// 1. Validate the draft and LLM availability.
	if text == "" {
		return quest.GrammarCheck{}, fmt.Errorf("text is required: %w", shared.ErrValidation)
	}
	if !uc.llm.Available() {
		return quest.GrammarCheck{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Ask the LLM with the lenient (non-strict) prompt.
	system, user := quest.BuildGrammarPrompts(text, "English", false)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return quest.GrammarCheck{}, err
	}

	// 3. Parse the verdict.
	var result quest.GrammarCheck
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return quest.GrammarCheck{}, fmt.Errorf("parse grammar check: %w", err)
	}
	result.OK = len(result.Errors) == 0
	return result, nil
}
