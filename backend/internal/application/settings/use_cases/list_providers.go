package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
)

type ListProvidersUseCase struct {
	repo settings.AIProviderRepository
}

func NewListProvidersUseCase(repo settings.AIProviderRepository) *ListProvidersUseCase {
	return &ListProvidersUseCase{repo: repo}
}

func (uc *ListProvidersUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]settings.AIProvider, error) {
	// 1. Only a global administrator manages platform settings.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return nil, err
	}
	// 2. Return the saved providers.
	return uc.repo.ListAIProviders(ctx)
}
