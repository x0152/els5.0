package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/reading/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	GenerateText  *usecases.GenerateTextUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "readingGenerateText",
		Method:      http.MethodPost,
		Path:        "/api/v1/reading/texts",
		Summary:     "Generate a one-page reading text",
		Tags:        []string{"reading"},
	}, func(ctx context.Context, actor *iam.Actor, in *GenerateTextInput) (TextOutput, error) {
		cmd, err := toGenerateTextCommand(in)
		if err != nil {
			return TextOutput{}, err
		}
		res, err := deps.GenerateText.Execute(ctx, actor, cmd)
		if err != nil {
			return TextOutput{}, err
		}
		words := res.Words
		if words == nil {
			words = []string{}
		}
		return TextOutput{Title: res.Title, Body: res.Body, Words: words}, nil
	})
}
