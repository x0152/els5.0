package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type AddUnitUseCase struct {
	units vocab.Repository
	llm   LLMClient
}

func NewAddUnitUseCase(units vocab.Repository, llm LLMClient) *AddUnitUseCase {
	return &AddUnitUseCase{units: units, llm: llm}
}

type AddUnitResult struct {
	Correct     bool
	Correction  string
	Explanation string
	Unit        *vocab.Unit
}

func (uc *AddUnitUseCase) Execute(ctx context.Context, actor *iam.Actor, input string) (AddUnitResult, error) {
	if actor == nil {
		return AddUnitResult{}, shared.ErrUnauthorized
	}
	accountID := actor.AccountID().String()

	// 1. Normalize the input.
	input = strings.TrimSpace(input)
	if input == "" {
		return AddUnitResult{}, shared.Validation(fmt.Errorf("text: must not be empty"))
	}
	if !uc.llm.Available() {
		return AddUnitResult{}, shared.ErrUnavailable
	}

	// 2. Ask the LLM to validate and describe the word.
	system, user := vocab.BuildCheckPrompt(input, actor.Account().NativeLanguage())
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return AddUnitResult{}, err
	}
	check, err := vocab.ParseCheckResult(raw)
	if err != nil {
		return AddUnitResult{}, err
	}

	// 3. If the input is invalid — return a correction and save nothing.
	if !check.Correct {
		return AddUnitResult{Correct: false, Correction: check.Correction, Explanation: check.Explanation}, nil
	}

	// 4. Skip duplicates in the user's collection.
	exists, err := uc.units.ExistsText(ctx, accountID, strings.TrimSpace(check.Text))
	if err != nil {
		return AddUnitResult{}, err
	}
	if exists {
		return AddUnitResult{}, shared.ErrConflict
	}

	// 5. Build the unit (domain validates invariants) and persist it.
	unit, err := vocab.NewUnit(uuid.NewString(), accountID, check)
	if err != nil {
		return AddUnitResult{}, err
	}
	stored, err := uc.units.Create(ctx, unit)
	if err != nil {
		return AddUnitResult{}, err
	}
	return AddUnitResult{Correct: true, Unit: &stored}, nil
}
