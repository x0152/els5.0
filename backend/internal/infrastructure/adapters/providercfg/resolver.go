package providercfg

import (
	"context"

	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
)

type Resolver struct {
	repo     settings.AIProviderRepository
	feature  settings.Feature
	fallback ports.AIProviderConfig
}

func NewResolver(repo settings.AIProviderRepository, feature settings.Feature, fallback ports.AIProviderConfig) *Resolver {
	return &Resolver{repo: repo, feature: feature, fallback: fallback}
}

func (r *Resolver) Resolve(ctx context.Context) ports.AIProviderConfig {
	provider, err := r.repo.GetAIProvider(ctx, r.feature)
	if err != nil {
		return r.fallback
	}
	cfg := ports.AIProviderConfig{
		Kind:    string(provider.Kind),
		BaseURL: provider.BaseURL,
		APIKey:  provider.APIKey,
		Model:   provider.Model,
		Params:  provider.Params,
	}
	if cfg.IsEmpty() {
		return r.fallback
	}
	return cfg
}
