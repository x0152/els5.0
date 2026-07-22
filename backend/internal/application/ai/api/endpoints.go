package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/ai/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	Service       *usecases.Service
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "aiHistory",
		Method:      http.MethodGet,
		Path:        "/api/v1/ai/history",
		Summary:     "Get current chat history and selected model",
		Tags:        []string{"ai"},
	}, func(ctx context.Context, actor *iam.Actor, _ *HistoryInput) (HistoryOutput, error) {
		res, err := deps.Service.History(ctx, actor)
		if err != nil {
			return HistoryOutput{}, err
		}
		return toHistoryOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "aiModels",
		Method:      http.MethodGet,
		Path:        "/api/v1/ai/models",
		Summary:     "List available chat models",
		Tags:        []string{"ai"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ModelsInput) (ModelsOutput, error) {
		res, err := deps.Service.Models(ctx, actor)
		if err != nil {
			return ModelsOutput{}, err
		}
		return toModelsOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "aiSetModel",
		Method:      http.MethodPost,
		Path:        "/api/v1/ai/model",
		Summary:     "Select chat model",
		Tags:        []string{"ai"},
	}, func(ctx context.Context, actor *iam.Actor, in *SetModelInput) (OKOutput, error) {
		if strings.TrimSpace(in.Body.Model) == "" {
			return OKOutput{}, huma.Error400BadRequest("model is required")
		}
		if err := deps.Service.SetModel(ctx, actor, in.Body.Model); err != nil {
			return OKOutput{}, err
		}
		return OKOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "aiResetContext",
		Method:      http.MethodPost,
		Path:        "/api/v1/ai/reset",
		Summary:     "Reset conversation context",
		Tags:        []string{"ai"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ResetInput) (OKOutput, error) {
		if err := deps.Service.Reset(ctx, actor); err != nil {
			return OKOutput{}, err
		}
		return OKOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "aiFillGap",
		Method:      http.MethodPost,
		Path:        "/api/v1/ai/fill-gap",
		Summary:     "Store the user's answer inside an exercise message",
		Tags:        []string{"ai"},
	}, func(ctx context.Context, actor *iam.Actor, in *FillGapInput) (OKOutput, error) {
		if err := deps.Service.FillGap(ctx, actor, in.Body.MessageID, in.Body.Ordinal, in.Body.Answer); err != nil {
			return OKOutput{}, err
		}
		return OKOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "aiClearChat",
		Method:      http.MethodPost,
		Path:        "/api/v1/ai/clear",
		Summary:     "Clear chat history",
		Tags:        []string{"ai"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ClearInput) (OKOutput, error) {
		if err := deps.Service.Clear(ctx, actor); err != nil {
			return OKOutput{}, err
		}
		return OKOutput{OK: true}, nil
	})
}
