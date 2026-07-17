package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/diary/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	GetToday      *usecases.GetTodayUseCase
	SubmitEntry   *usecases.SubmitEntryUseCase
	ListEntries   *usecases.ListEntriesUseCase
	ResetHistory  *usecases.ResetHistoryUseCase
	TrainerCheck  *usecases.TrainerCheckUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "diaryToday",
		Method:      http.MethodGet,
		Path:        "/api/v1/diary/today",
		Summary:     "Today's diary state: question, warmup and streak",
		Tags:        []string{"diary"},
	}, func(ctx context.Context, actor *iam.Actor, _ *GetTodayInput) (TodayOutput, error) {
		res, err := deps.GetToday.Execute(ctx, actor)
		if err != nil {
			return TodayOutput{}, err
		}
		return toTodayOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "diarySubmitEntry",
		Method:      http.MethodPost,
		Path:        "/api/v1/diary/entries",
		Summary:     "Submit today's entry and get the friend reply with corrections",
		Tags:        []string{"diary"},
	}, func(ctx context.Context, actor *iam.Actor, in *SubmitEntryInput) (EntryOutput, error) {
		res, err := deps.SubmitEntry.Execute(ctx, actor, toSubmitEntryCommand(in))
		if err != nil {
			return EntryOutput{}, err
		}
		return toEntryOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "diaryListEntries",
		Method:      http.MethodGet,
		Path:        "/api/v1/diary/entries",
		Summary:     "List past diary entries",
		Tags:        []string{"diary"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListEntriesInput) (EntriesOutput, error) {
		res, err := deps.ListEntries.Execute(ctx, actor, toListEntriesQuery(in))
		if err != nil {
			return EntriesOutput{}, err
		}
		return toEntriesOutput(res, in.Limit, in.Offset), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "diaryResetHistory",
		Method:        http.MethodDelete,
		Path:          "/api/v1/diary/entries",
		Summary:       "Delete all diary entries of the account",
		Tags:          []string{"diary"},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, actor *iam.Actor, _ *ResetHistoryInput) (ResetHistoryOutput, error) {
		return ResetHistoryOutput{}, deps.ResetHistory.Execute(ctx, actor)
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "diaryTrainerCheck",
		Method:      http.MethodPost,
		Path:        "/api/v1/diary/trainer/check",
		Summary:     "Check a draft reply without revealing corrections",
		Tags:        []string{"diary"},
	}, func(ctx context.Context, actor *iam.Actor, in *TrainerCheckInput) (TrainerCheckOutput, error) {
		cmd, err := toTrainerCheckCommand(in)
		if err != nil {
			return TrainerCheckOutput{}, err
		}
		res, err := deps.TrainerCheck.Execute(ctx, actor, cmd)
		if err != nil {
			return TrainerCheckOutput{}, err
		}
		return toTrainerCheckOutput(res), nil
	})
}
