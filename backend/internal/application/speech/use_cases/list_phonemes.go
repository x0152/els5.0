package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/speech"
)

type ListPhonemesUseCase struct{}

func NewListPhonemesUseCase() *ListPhonemesUseCase {
	return &ListPhonemesUseCase{}
}

func (uc *ListPhonemesUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]speech.PhonemeInfo, error) {
	// 1. Return the static articulation guide.
	return speech.PhonemeGuide, nil
}
