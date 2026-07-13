package usecases

import (
	"context"
	"errors"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared"
)

type UpdateProviderCommand struct {
	Feature     settings.Feature
	BaseURL     string
	Model       string
	APIKey      string
	KeyProvided bool
}

type UpdateProviderUseCase struct {
	repo settings.AIProviderRepository
}

func NewUpdateProviderUseCase(repo settings.AIProviderRepository) *UpdateProviderUseCase {
	return &UpdateProviderUseCase{repo: repo}
}

func (uc *UpdateProviderUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd UpdateProviderCommand) (settings.AIProvider, error) {
	// 1. Only a global administrator manages platform settings.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return settings.AIProvider{}, err
	}
	// 2. Keep the current token if a new one was not provided (write-only field).
	apiKey := cmd.APIKey
	if !cmd.KeyProvided {
		existing, err := uc.repo.GetAIProvider(ctx, cmd.Feature)
		if err != nil && !errors.Is(err, shared.ErrNotFound) {
			return settings.AIProvider{}, err
		}
		apiKey = existing.APIKey
	}
	// 3. Overwrite the provider.
	provider := settings.NewAIProvider(cmd.Feature, cmd.BaseURL, apiKey, cmd.Model)
	if err := uc.repo.UpsertAIProvider(ctx, provider); err != nil {
		return settings.AIProvider{}, err
	}
	return provider, nil
}
