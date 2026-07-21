package ports

import "context"

type AIProviderConfig struct {
	Kind    string
	BaseURL string
	APIKey  string
	Model   string
	Params  map[string]string
}

func (c AIProviderConfig) IsEmpty() bool {
	return c.BaseURL == "" && c.APIKey == "" && c.Model == ""
}

type AIProviderResolver interface {
	Resolve(ctx context.Context) AIProviderConfig
}

type ModelLister interface {
	ListModels(ctx context.Context, cfg AIProviderConfig) ([]string, error)
}
