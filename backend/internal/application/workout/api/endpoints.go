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
	Authenticator  *authx.Authenticator
	GetToday       *usecases.GetTodayUseCase
	GenerateLesson *usecases.GenerateLessonUseCase
	GetLesson      *usecases.GetLessonUseCase
	SubmitStep     *usecases.SubmitStepUseCase
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
		Summary:     "Return the active lesson, generating one if none is ready",
		Tags:        []string{"workout"},
	}, func(ctx context.Context, actor *iam.Actor, _ *StartLessonInput) (LessonOutput, error) {
		lesson, err := deps.GenerateLesson.Execute(ctx, actor.AccountID().String())
		if err != nil {
			return LessonOutput{}, err
		}
		return toLessonOutput(lesson), nil
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
}
