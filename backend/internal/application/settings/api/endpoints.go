package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/settings/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator      *authx.Authenticator
	ListProviders      *usecases.ListProvidersUseCase
	UpdateProvider     *usecases.UpdateProviderUseCase
	ListModels         *usecases.ListModelsUseCase
	GetEventProcessing *usecases.GetFlagUseCase
	SetEventProcessing *usecases.SetFlagUseCase
	GetAutoWordImages  *usecases.GetFlagUseCase
	SetAutoWordImages  *usecases.SetFlagUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listAIProviders",
		Method:      http.MethodGet,
		Path:        "/api/v1/settings/ai/providers",
		Summary:     "List AI provider settings for every platform feature",
		Tags:        []string{"settings"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListProvidersInput) (ProvidersOutput, error) {
		list, err := deps.ListProviders.Execute(ctx, actor)
		if err != nil {
			return ProvidersOutput{}, err
		}
		return toProvidersOutput(list), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "updateAIProvider",
		Method:      http.MethodPut,
		Path:        "/api/v1/settings/ai/providers/{feature}",
		Summary:     "Update base URL, token and model for an AI provider",
		Tags:        []string{"settings"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateProviderInput) (ProviderResponse, error) {
		cmd, err := toUpdateProviderCommand(in)
		if err != nil {
			return ProviderResponse{}, err
		}
		provider, err := deps.UpdateProvider.Execute(ctx, actor, cmd)
		if err != nil {
			return ProviderResponse{}, err
		}
		return ProviderResponse{Provider: toProviderOutput(provider)}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listAIProviderModels",
		Method:      http.MethodGet,
		Path:        "/api/v1/settings/ai/providers/{feature}/models",
		Summary:     "List models offered by an AI provider endpoint",
		Tags:        []string{"settings"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListProviderModelsInput) (ProviderModelsOutput, error) {
		feature, err := settings.ParseFeature(in.Feature)
		if err != nil {
			return ProviderModelsOutput{}, err
		}
		override := ports.AIProviderConfig{BaseURL: in.BaseURL, APIKey: in.APIKey}
		models, err := deps.ListModels.Execute(ctx, actor, feature, override)
		if err != nil {
			return ProviderModelsOutput{}, err
		}
		return toModelsOutput(models), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getEventProcessing",
		Method:      http.MethodGet,
		Path:        "/api/v1/settings/event-processing",
		Summary:     "Whether pending events are processed by workers",
		Tags:        []string{"settings"},
	}, func(ctx context.Context, actor *iam.Actor, _ *EventProcessingInput) (EventProcessingOutput, error) {
		enabled, err := deps.GetEventProcessing.Execute(ctx, actor)
		if err != nil {
			return EventProcessingOutput{}, err
		}
		return EventProcessingOutput{Enabled: enabled}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "setEventProcessing",
		Method:      http.MethodPut,
		Path:        "/api/v1/settings/event-processing",
		Summary:     "Enable or disable processing of pending events",
		Tags:        []string{"settings"},
	}, func(ctx context.Context, actor *iam.Actor, in *SetEventProcessingInput) (EventProcessingOutput, error) {
		if err := deps.SetEventProcessing.Execute(ctx, actor, in.Body.Enabled); err != nil {
			return EventProcessingOutput{}, err
		}
		return EventProcessingOutput{Enabled: in.Body.Enabled}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getAutoWordImages",
		Method:      http.MethodGet,
		Path:        "/api/v1/settings/auto-word-images",
		Summary:     "Whether illustrations are generated automatically for new vocabulary words",
		Tags:        []string{"settings"},
	}, func(ctx context.Context, actor *iam.Actor, _ *EventProcessingInput) (EventProcessingOutput, error) {
		enabled, err := deps.GetAutoWordImages.Execute(ctx, actor)
		if err != nil {
			return EventProcessingOutput{}, err
		}
		return EventProcessingOutput{Enabled: enabled}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "setAutoWordImages",
		Method:      http.MethodPut,
		Path:        "/api/v1/settings/auto-word-images",
		Summary:     "Enable or disable automatic illustration generation for new vocabulary words",
		Tags:        []string{"settings"},
	}, func(ctx context.Context, actor *iam.Actor, in *SetEventProcessingInput) (EventProcessingOutput, error) {
		if err := deps.SetAutoWordImages.Execute(ctx, actor, in.Body.Enabled); err != nil {
			return EventProcessingOutput{}, err
		}
		return EventProcessingOutput{Enabled: in.Body.Enabled}, nil
	})
}
