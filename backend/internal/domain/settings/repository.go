package settings

import "context"

type AIProviderRepository interface {
	ListAIProviders(ctx context.Context) ([]AIProvider, error)
	GetAIProvider(ctx context.Context, feature Feature) (AIProvider, error)
	UpsertAIProvider(ctx context.Context, provider AIProvider) error
}
