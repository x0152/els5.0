package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type ProviderOutput struct {
	Feature string `json:"feature"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	HasKey  bool   `json:"has_key"`
}

type ProvidersOutput struct {
	Items []ProviderOutput `json:"items"`
}

type ListProvidersInput struct {
	authx.BearerInput
}

type UpdateProviderInput struct {
	authx.BearerInput
	Feature string `path:"feature" enum:"main,analysis,vision,image"`
	Body    struct {
		BaseURL string  `json:"base_url" maxLength:"500" doc:"Provider base URL (OpenAI-compatible)"`
		Model   string  `json:"model" maxLength:"200" doc:"Model id from the provider /models list"`
		APIKey  *string `json:"api_key,omitempty" maxLength:"500" doc:"API token; omit to keep the current one"`
	}
}

type ProviderResponse struct {
	Provider ProviderOutput `json:"provider"`
}

type ListProviderModelsInput struct {
	authx.BearerInput
	Feature string `path:"feature" enum:"main,analysis,vision,image"`
	BaseURL string `query:"base_url" maxLength:"500" doc:"Override base URL to query instead of the saved one"`
	APIKey  string `query:"api_key" maxLength:"500" doc:"Override API token; omit to reuse the saved one"`
}

type ProviderModelsOutput struct {
	Items []string `json:"items"`
}

type EventProcessingInput struct {
	authx.BearerInput
}

type SetEventProcessingInput struct {
	authx.BearerInput
	Body struct {
		Enabled bool `json:"enabled" doc:"Process pending events when true; keep them pending when false"`
	}
}

type EventProcessingOutput struct {
	Enabled bool `json:"enabled"`
}
