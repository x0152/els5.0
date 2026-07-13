package usecases

import (
	"context"
	"errors"

	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
)

type SeedProvidersUseCase struct {
	repo     settings.AIProviderRepository
	defaults map[settings.Feature]ports.AIProviderConfig
}

func NewSeedProvidersUseCase(repo settings.AIProviderRepository, defaults map[settings.Feature]ports.AIProviderConfig) *SeedProvidersUseCase {
	return &SeedProvidersUseCase{repo: repo, defaults: defaults}
}

func (uc *SeedProvidersUseCase) Execute(ctx context.Context) error {
	// 1. For each feature create a row from env if it is not yet in the DB.
	for _, feature := range settings.Features() {
		_, err := uc.repo.GetAIProvider(ctx, feature)
		if err == nil {
			continue
		}
		if !errors.Is(err, shared.ErrNotFound) {
			return err
		}
		d := uc.defaults[feature]
		if err := uc.repo.UpsertAIProvider(ctx, settings.NewAIProvider(feature, d.BaseURL, d.APIKey, d.Model)); err != nil {
			return err
		}
	}
	return nil
}
