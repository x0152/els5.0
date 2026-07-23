package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/onboarding/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	GetProgress   *usecases.GetProgressUseCase
	AckItems      *usecases.AckItemsUseCase
	GetTours      *usecases.GetToursUseCase
	MarkTour      *usecases.MarkTourUseCase
	ResetTours    *usecases.ResetToursUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "onboardingProgress",
		Method:      http.MethodGet,
		Path:        "/api/v1/onboarding/progress",
		Summary:     "Getting-started checklist and achievements with accumulated progress",
		Tags:        []string{"onboarding"},
	}, func(ctx context.Context, actor *iam.Actor, _ *GetProgressInput) (OnboardingProgressOutput, error) {
		res, err := deps.GetProgress.Execute(ctx, actor)
		if err != nil {
			return OnboardingProgressOutput{}, err
		}
		return toProgressOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "onboardingAck",
		Method:        http.MethodPost,
		Path:          "/api/v1/onboarding/ack",
		Summary:       "Mark unlocked achievements as seen",
		Tags:          []string{"onboarding"},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, actor *iam.Actor, in *AckItemsInput) (AckItemsOutput, error) {
		return AckItemsOutput{}, deps.AckItems.Execute(ctx, actor, in.Body.IDs)
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "onboardingTours",
		Method:      http.MethodGet,
		Path:        "/api/v1/onboarding/tours",
		Summary:     "Completed onboarding tour ids",
		Tags:        []string{"onboarding"},
	}, func(ctx context.Context, actor *iam.Actor, _ *GetToursInput) (ToursOutput, error) {
		ids, err := deps.GetTours.Execute(ctx, actor)
		if err != nil {
			return ToursOutput{}, err
		}
		return toToursOutput(ids), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "onboardingMarkTour",
		Method:        http.MethodPost,
		Path:          "/api/v1/onboarding/tours",
		Summary:       "Mark an onboarding tour as completed",
		Tags:          []string{"onboarding"},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, actor *iam.Actor, in *MarkTourInput) (MarkTourOutput, error) {
		return MarkTourOutput{}, deps.MarkTour.Execute(ctx, actor, in.Body.ID)
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "onboardingResetTours",
		Method:        http.MethodDelete,
		Path:          "/api/v1/onboarding/tours",
		Summary:       "Reset onboarding tours so they show again",
		Tags:          []string{"onboarding"},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, actor *iam.Actor, _ *ResetToursInput) (ResetToursOutput, error) {
		return ResetToursOutput{}, deps.ResetTours.Execute(ctx, actor)
	})
}
