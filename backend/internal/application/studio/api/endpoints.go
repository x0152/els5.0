package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/studio/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/studio"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	ListAreas     *usecases.ListAreasUseCase
	CreateArea    *usecases.CreateAreaUseCase
	DeleteArea    *usecases.DeleteAreaUseCase
	ListItems     *usecases.ListItemsUseCase
	AddItem       *usecases.AddItemUseCase
	CaptureItem   *usecases.CaptureItemUseCase
	DeleteItem    *usecases.DeleteItemUseCase
	MarkSkill     *usecases.MarkSkillUseCase
	PassReview    *usecases.PassReviewUseCase
	RegenExample  *usecases.RegenExampleUseCase
	RegenTask     *usecases.RegenTaskUseCase
	CheckReply    *usecases.CheckReplyUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioListAreas",
		Method:      http.MethodGet,
		Path:        "/api/v1/studio/areas",
		Summary:     "List study areas with progress",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListAreasInput) (AreasOutput, error) {
		res, err := deps.ListAreas.Execute(ctx, actor)
		if err != nil {
			return AreasOutput{}, err
		}
		return toAreasOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioCreateArea",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/areas",
		Summary:     "Create a study area",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *CreateAreaInput) (AreaOutput, error) {
		res, err := deps.CreateArea.Execute(ctx, actor, usecases.CreateAreaCommand{Title: in.Body.Title, Icon: in.Body.Icon})
		if err != nil {
			return AreaOutput{}, err
		}
		return toAreaOutput(studio.AreaStats{Area: res}), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "studioDeleteArea",
		Method:        http.MethodDelete,
		Path:          "/api/v1/studio/areas/{id}",
		Summary:       "Delete a study area with its items",
		Tags:          []string{"studio"},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteAreaInput) (DeleteAreaOutput, error) {
		return DeleteAreaOutput{}, deps.DeleteArea.Execute(ctx, actor, in.ID)
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioListItems",
		Method:      http.MethodGet,
		Path:        "/api/v1/studio/areas/{id}/items",
		Summary:     "List items of a study area",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListItemsInput) (ItemsOutput, error) {
		res, err := deps.ListItems.Execute(ctx, actor, in.ID)
		if err != nil {
			return ItemsOutput{}, err
		}
		return toItemsOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioAddItem",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/areas/{id}/items",
		Summary:     "Add a phrase or word to a study area (AI fills transcription, translation and example)",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *AddItemInput) (ItemOutput, error) {
		res, err := deps.AddItem.Execute(ctx, actor, usecases.AddItemCommand{AreaID: in.ID, Text: in.Body.Text})
		if err != nil {
			return ItemOutput{}, err
		}
		return toItemOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioCaptureItem",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/capture",
		Summary:     "Add a phrase to a named area (created if missing) — used by other apps",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *CaptureItemInput) (ItemOutput, error) {
		res, err := deps.CaptureItem.Execute(ctx, actor, usecases.CaptureItemCommand{
			Text:      in.Body.Text,
			AreaTitle: in.Body.Area,
			Icon:      in.Body.Icon,
		})
		if err != nil {
			return ItemOutput{}, err
		}
		return toItemOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "studioDeleteItem",
		Method:        http.MethodDelete,
		Path:          "/api/v1/studio/items/{id}",
		Summary:       "Delete a study item",
		Tags:          []string{"studio"},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteItemInput) (DeleteItemOutput, error) {
		return DeleteItemOutput{}, deps.DeleteItem.Execute(ctx, actor, in.ID)
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioMarkSkill",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/items/{id}/skill",
		Summary:     "Mark a skill (listened/spoken/written/recalled) as done for an item",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *MarkSkillInput) (ItemOutput, error) {
		res, err := deps.MarkSkill.Execute(ctx, actor, usecases.MarkSkillCommand{ItemID: in.ID, Skill: in.Body.Skill})
		if err != nil {
			return ItemOutput{}, err
		}
		return toItemOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioPassReview",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/items/{id}/review",
		Summary:     "Pass the due review of an item and schedule the next one",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *PassReviewInput) (ItemOutput, error) {
		res, err := deps.PassReview.Execute(ctx, actor, in.ID)
		if err != nil {
			return ItemOutput{}, err
		}
		return toItemOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioRegenExample",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/items/{id}/example",
		Summary:     "Regenerate the usage example for an item",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *RegenExampleInput) (ItemOutput, error) {
		res, err := deps.RegenExample.Execute(ctx, actor, in.ID)
		if err != nil {
			return ItemOutput{}, err
		}
		return toItemOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioRegenTask",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/items/{id}/task",
		Summary:     "Generate or regenerate the 'use it' mini-situation for an item",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *RegenTaskInput) (ItemOutput, error) {
		res, err := deps.RegenTask.Execute(ctx, actor, in.ID)
		if err != nil {
			return ItemOutput{}, err
		}
		return toItemOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "studioCheckReply",
		Method:      http.MethodPost,
		Path:        "/api/v1/studio/items/{id}/check",
		Summary:     "Check the user's reply to the 'use it' task; success marks the written skill",
		Tags:        []string{"studio"},
	}, func(ctx context.Context, actor *iam.Actor, in *CheckReplyInput) (CheckReplyOutput, error) {
		res, err := deps.CheckReply.Execute(ctx, actor, usecases.CheckReplyCommand{ItemID: in.ID, Reply: in.Body.Reply})
		if err != nil {
			return CheckReplyOutput{}, err
		}
		return toCheckReplyOutput(res), nil
	})
}
