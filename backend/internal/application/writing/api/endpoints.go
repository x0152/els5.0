package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/writing/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator     *authx.Authenticator
	TrainerCheck      *usecases.TrainerCheckUseCase
	GenerateSituation *usecases.GenerateSituationUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "writingTrainerCheck",
		Method:      http.MethodPost,
		Path:        "/api/v1/writing/trainer/check",
		Summary:     "Check a draft reply without revealing corrections",
		Tags:        []string{"writing"},
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

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "writingGenerateSituation",
		Method:      http.MethodPost,
		Path:        "/api/v1/writing/trainer/situations",
		Summary:     "Generate a dialogue situation to reply to",
		Tags:        []string{"writing"},
	}, func(ctx context.Context, actor *iam.Actor, in *GenerateSituationInput) (SituationOutput, error) {
		res, err := deps.GenerateSituation.Execute(ctx, actor, usecases.GenerateSituationCommand{Topic: in.Body.Topic})
		if err != nil {
			return SituationOutput{}, err
		}
		return SituationOutput{Scenario: res.Scenario, Dialogue: res.Dialogue}, nil
	})
}
