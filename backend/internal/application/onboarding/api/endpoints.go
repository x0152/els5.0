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
}
