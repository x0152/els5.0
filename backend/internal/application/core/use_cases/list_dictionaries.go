package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/iam"
)

type ListDictionariesUseCase struct{}

func NewListDictionariesUseCase() *ListDictionariesUseCase {
	return &ListDictionariesUseCase{}
}

func (uc *ListDictionariesUseCase) Execute(ctx context.Context, actor *iam.Actor) (map[string][]core.DictEntry, error) {
	return core.Dictionaries(), nil
}
