package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/workout/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	GetToday      *usecases.GetTodayUseCase
	StartLesson   *usecases.StartLessonUseCase
	GetLesson     *usecases.GetLessonUseCase
	SubmitStep    *usecases.SubmitStepUseCase
	Reset         *usecases.ResetUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "workoutToday",
		Method:      http.MethodGet,
		Path:        "/api/v1/workout/today",
		Summary:     "Current lesson, streak and completion calendar",
		Tags:        []string{"workout"},
	}, func(ctx context.Context, actor *iam.Actor, _ *TodayInput) (WorkoutTodayOutput, error) {
		res, err := deps.GetToday.Execute(ctx, actor)
		if err != nil {
			return WorkoutTodayOutput{}, err
		}
		return toTodayOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "workoutStartLesson",
		Method:      http.MethodPost,
		Path:        "/api/v1/workout/lessons",
		Summary:     "Return the active lesson or start background generation",
		Tags:        []string{"workout"},
	}, func(ctx context.Context, actor *iam.Actor, _ *StartLessonInput) (StartLessonOutput, error) {
		res, err := deps.StartLesson.Execute(ctx, actor.AccountID().String())
		if err != nil {
			return StartLessonOutput{}, err
		}
		return toStartLessonOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "workoutGetLesson",
		Method:      http.MethodGet,
		Path:        "/api/v1/workout/lessons/{id}",
		Summary:     "Get a lesson with all step payloads",
		Tags:        []string{"workout"},
	}, func(ctx context.Context, actor *iam.Actor, in *GetLessonInput) (LessonOutput, error) {
		lesson, err := deps.GetLesson.Execute(ctx, actor, in.ID)
		if err != nil {
			return LessonOutput{}, err
		}
		return toLessonOutput(lesson), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "workoutSubmitStep",
		Method:      http.MethodPost,
		Path:        "/api/v1/workout/lessons/{id}/steps/{step}",
		Summary:     "Submit a step result; the lesson completes with its last step",
		Tags:        []string{"workout"},
	}, func(ctx context.Context, actor *iam.Actor, in *SubmitStepInput) (LessonOutput, error) {
		lesson, err := deps.SubmitStep.Execute(ctx, actor, toSubmitStepCommand(in))
		if err != nil {
			return LessonOutput{}, err
		}
		return toLessonOutput(lesson), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "workoutReset",
		Method:        http.MethodDelete,
		Path:          "/api/v1/workout",
		Summary:       "Delete all workout progress and generated lessons of the account",
		Tags:          []string{"workout"},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, actor *iam.Actor, _ *ResetInput) (ResetOutput, error) {
		return ResetOutput{}, deps.Reset.Execute(ctx, actor)
	})
}
