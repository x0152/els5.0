package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/vocab"
	"github.com/els/backend/internal/utils/timex"
)

type AnswerCardUseCase struct {
	units vocab.Repository
	clock timex.Clock
}

func NewAnswerCardUseCase(units vocab.Repository, clock timex.Clock) *AnswerCardUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &AnswerCardUseCase{units: units, clock: clock}
}

type AnswerCardResult struct {
	Correct bool
	Unit    vocab.Unit
}

func (uc *AnswerCardUseCase) Execute(ctx context.Context, actor *iam.Actor, unitID, answer string) (AnswerCardResult, error) {
	// 1. Load the unit.
	unit, err := uc.units.Get(ctx, actor.AccountID().String(), unitID)
	if err != nil {
		return AnswerCardResult{}, err
	}

	// 2. Check the answer and apply streak/status progression.
	correct := vocab.IsCorrectAnswer(unit, answer)
	updated := vocab.ApplyAnswer(unit, correct, uc.clock.Now())

	// 3. Persist progress.
	saved, err := uc.units.UpdateProgress(ctx, updated)
	if err != nil {
		return AnswerCardResult{}, err
	}
	return AnswerCardResult{Correct: correct, Unit: saved}, nil
}
