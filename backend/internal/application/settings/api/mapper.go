package api

import (
	usecases "github.com/els/backend/internal/application/settings/use_cases"
	"github.com/els/backend/internal/domain/settings"
)

func toProviderOutput(p settings.AIProvider) ProviderOutput {
	kind := p.Kind
	if kind == "" {
		kind = settings.KindOpenAI
	}
	params := p.Params
	if params == nil {
		params = map[string]string{}
	}
	return ProviderOutput{
		Feature: string(p.Feature),
		Kind:    string(kind),
		BaseURL: p.BaseURL,
		Model:   p.Model,
		HasKey:  p.HasKey(),
		Params:  params,
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
	kind, err := settings.ParseKind(in.Body.Kind)
	if err != nil {
		return usecases.UpdateProviderCommand{}, err
	}
	cmd := usecases.UpdateProviderCommand{
		Feature: feature,
		Kind:    kind,
		BaseURL: in.Body.BaseURL,
		Model:   in.Body.Model,
		Params:  in.Body.Params,
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
