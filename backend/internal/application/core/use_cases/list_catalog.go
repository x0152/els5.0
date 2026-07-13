package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/iam"
)

type ListCatalogUseCase struct {
	store EventStore
}

func NewListCatalogUseCase(store EventStore) *ListCatalogUseCase {
	return &ListCatalogUseCase{store: store}
}

type ListCatalogResult struct {
	Words []core.Word
	Rules []core.GrammarRule
}

func (uc *ListCatalogUseCase) Execute(ctx context.Context, actor *iam.Actor) (ListCatalogResult, error) {
	// 1. Return recently updated catalog words.
	words, err := uc.store.ListWords(ctx, 500)
	if err != nil {
		return ListCatalogResult{}, err
	}

	// 2. And grammar rules.
	rules, err := uc.store.ListGrammarRules(ctx, 500)
	if err != nil {
		return ListCatalogResult{}, err
	}

	return ListCatalogResult{Words: words, Rules: rules}, nil
}
