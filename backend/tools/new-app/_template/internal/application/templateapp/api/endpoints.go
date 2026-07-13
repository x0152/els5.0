package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/templateapp/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	Echo          *usecases.EchoUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "echoTemplateapp",
		Method:      http.MethodPost,
		Path:        "/api/v1/templateapp/echo",
		Summary:     "Echo message (template module)",
		Tags:        []string{"templateapp"},
	}, func(ctx context.Context, actor *iam.Actor, in *EchoInput) (EchoOutput, error) {
		res, err := deps.Echo.Execute(ctx, actor, usecases.EchoCommand{Message: in.Body.Message})
		if err != nil {
			return EchoOutput{}, err
		}
		return EchoOutput{Message: res.Message}, nil
	})
}
