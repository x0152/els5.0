package usecases

import (
	"context"
	"errors"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
)

type ListModelsUseCase struct {
	repo     settings.AIProviderRepository
	lister   ports.ModelLister
	defaults map[settings.Feature]ports.AIProviderConfig
}

func NewListModelsUseCase(repo settings.AIProviderRepository, lister ports.ModelLister, defaults map[settings.Feature]ports.AIProviderConfig) *ListModelsUseCase {
	return &ListModelsUseCase{repo: repo, lister: lister, defaults: defaults}
}

func (uc *ListModelsUseCase) Execute(ctx context.Context, actor *iam.Actor, feature settings.Feature, override ports.AIProviderConfig) ([]string, error) {
	// 1. Only a global administrator manages platform settings.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return nil, err
	}
	// 2. Take the saved provider config, otherwise the env default.
	provider, err := uc.repo.GetAIProvider(ctx, feature)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}
	cfg := ports.AIProviderConfig{BaseURL: provider.BaseURL, APIKey: provider.APIKey, Model: provider.Model}
	if cfg.IsEmpty() {
		cfg = uc.defaults[feature]
	}
	// 3. Override address and token with form values if provided.
	if override.BaseURL != "" {
		cfg.BaseURL = override.BaseURL
	}
	if override.APIKey != "" {
		cfg.APIKey = override.APIKey
	}
	// 4. Ask the provider for the model list.
	models, err := uc.lister.ListModels(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return models, nil
}
