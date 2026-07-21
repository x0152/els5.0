package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/listening/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator     *authx.Authenticator
	GenerateDictation *usecases.GenerateDictationUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listeningGenerateDictation",
		Method:      http.MethodPost,
		Path:        "/api/v1/listening/dictations",
		Summary:     "Generate dictation sentences",
		Tags:        []string{"listening"},
	}, func(ctx context.Context, actor *iam.Actor, in *GenerateDictationInput) (DictationOutput, error) {
		cmd, err := toGenerateDictationCommand(in)
		if err != nil {
			return DictationOutput{}, err
		}
		res, err := deps.GenerateDictation.Execute(ctx, actor, cmd)
		if err != nil {
			return DictationOutput{}, err
		}
		return DictationOutput{Sentences: res.Sentences}, nil
	})
}
