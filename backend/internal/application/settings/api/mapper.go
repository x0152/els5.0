package api

import (
	usecases "github.com/els/backend/internal/application/settings/use_cases"
	"github.com/els/backend/internal/domain/settings"
)

func toProviderOutput(p settings.AIProvider) ProviderOutput {
	return ProviderOutput{
		Feature: string(p.Feature),
		BaseURL: p.BaseURL,
		Model:   p.Model,
		HasKey:  p.HasKey(),
	}
}

func toProvidersOutput(list []settings.AIProvider) ProvidersOutput {
	items := make([]ProviderOutput, 0, len(list))
	for _, p := range list {
		items = append(items, toProviderOutput(p))
	}
	return ProvidersOutput{Items: items}
}

func toUpdateProviderCommand(in *UpdateProviderInput) (usecases.UpdateProviderCommand, error) {
	feature, err := settings.ParseFeature(in.Feature)
	if err != nil {
		return usecases.UpdateProviderCommand{}, err
	}
	cmd := usecases.UpdateProviderCommand{
		Feature: feature,
		BaseURL: in.Body.BaseURL,
		Model:   in.Body.Model,
	}
	if in.Body.APIKey != nil {
		cmd.KeyProvided = true
		cmd.APIKey = *in.Body.APIKey
	}
	return cmd, nil
}

func toModelsOutput(models []string) ProviderModelsOutput {
	if models == nil {
		models = []string{}
	}
	return ProviderModelsOutput{Items: models}
}
