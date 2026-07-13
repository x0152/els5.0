package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/quest/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/quest"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator    *authx.Authenticator
	CreateMission    *usecases.CreateMissionUseCase
	ListMissions     *usecases.ListMissionsUseCase
	GetMission       *usecases.GetMissionUseCase
	StartRespond     *usecases.StartRespondUseCase
	SuggestNative    *usecases.SuggestNativeReplyUseCase
	ResetMission     *usecases.ResetMissionUseCase
	RegenerateImages *usecases.RegenerateImagesUseCase
	DeleteMission    *usecases.DeleteMissionUseCase
	MediaURLs        media.PublicURL
	MediaBucket      string
}

func Register(api huma.API, deps Deps) {
	urls := mediaURLs{urls: deps.MediaURLs, bucket: deps.MediaBucket}

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "createQuestMission",
		Method:      http.MethodPost,
		Path:        "/api/v1/quest/missions",
		Summary:     "Create a quest mission",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, in *CreateMissionInput) (CreateMissionOutput, error) {
		res, err := deps.CreateMission.Execute(ctx, actor, quest.CreateMissionRequest{
			Prompt:        in.Body.Prompt,
			Genre:         in.Body.Genre,
			Language:      in.Body.Language,
			PracticeGoals: in.Body.PracticeGoals,
		})
		var out CreateMissionOutput
		if err != nil {
			return out, err
		}
		out.MissionID = res.MissionID
		return out, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listQuestMissions",
		Method:      http.MethodGet,
		Path:        "/api/v1/quest/missions",
		Summary:     "List quest missions with generation status",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListMissionsInput) (ListMissionsOutput, error) {
		missions, err := deps.ListMissions.Execute(ctx, actor)
		var out ListMissionsOutput
		if err != nil {
			return out, err
		}
		out.Missions = make([]MissionSummary, 0, len(missions))
		for _, item := range missions {
			out.Missions = append(out.Missions, toMissionSummary(item, urls))
		}
		return out, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getQuestMission",
		Method:      http.MethodGet,
		Path:        "/api/v1/quest/missions/{id}",
		Summary:     "Get a quest mission (polling: generation, images, active reply)",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, in *MissionInput) (GetMissionOutput, error) {
		res, err := deps.GetMission.Execute(ctx, actor, in.ID)
		var out GetMissionOutput
		if err != nil {
			return out, err
		}
		out.Mission = toMissionView(res.Mission, urls)
		out.ActiveReply = res.ActiveReply
		return out, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "respondQuestMission",
		Method:      http.MethodPost,
		Path:        "/api/v1/quest/missions/{id}/respond",
		Summary:     "Submit a player response (async)",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, in *RespondInput) (RespondOutput, error) {
		res, err := deps.StartRespond.Execute(ctx, actor, in.ID, quest.RespondRequest{Text: in.Body.Text, Strict: in.Body.Strict})
		var out RespondOutput
		if err != nil {
			return out, err
		}
		out.JobID = res.JobID
		return out, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "suggestQuestNativeReply",
		Method:      http.MethodPost,
		Path:        "/api/v1/quest/missions/{id}/native-reply",
		Summary:     "Suggest native-like variants for the player reply",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, in *SuggestNativeReplyInput) (SuggestNativeReplyOutput, error) {
		res, err := deps.SuggestNative.Execute(ctx, actor, in.ID, in.Body.Text)
		var out SuggestNativeReplyOutput
		if err != nil {
			return out, err
		}
		out.Variants = res.Variants
		return out, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "resetQuestMission",
		Method:      http.MethodPost,
		Path:        "/api/v1/quest/missions/{id}/reset",
		Summary:     "Reset a quest mission to its first scene",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, in *MissionInput) (ResetMissionOutput, error) {
		mission, err := deps.ResetMission.Execute(ctx, actor, in.ID)
		var out ResetMissionOutput
		if err != nil {
			return out, err
		}
		out.Mission = toMissionView(mission, urls)
		return out, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "regenerateQuestMissionImages",
		Method:      http.MethodPost,
		Path:        "/api/v1/quest/missions/{id}/regenerate-images",
		Summary:     "Regenerate failed mission images (cover, scenes, avatars)",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, in *RegenerateImagesInput) (RegenerateImagesOutput, error) {
		mission, err := deps.RegenerateImages.Execute(ctx, actor, in.ID, usecases.RegenerateImagesCommand{Kind: in.Body.Kind, Key: in.Body.Key})
		var out RegenerateImagesOutput
		if err != nil {
			return out, err
		}
		out.Mission = toMissionView(mission, urls)
		return out, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "deleteQuestMission",
		Method:      http.MethodDelete,
		Path:        "/api/v1/quest/missions/{id}",
		Summary:     "Delete a quest mission",
		Tags:        []string{"quest"},
	}, func(ctx context.Context, actor *iam.Actor, in *MissionInput) (DeleteMissionOutput, error) {
		var out DeleteMissionOutput
		if err := deps.DeleteMission.Execute(ctx, actor, in.ID); err != nil {
			return out, err
		}
		out.OK = true
		return out, nil
	})
}
