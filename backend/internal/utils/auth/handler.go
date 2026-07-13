package auth

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/utils/httpx"
)

type BearerInput struct {
	Authorization string `header:"Authorization" doc:"Bearer <token>"`
}

func (b BearerInput) GetAuthorization() string { return b.Authorization }

type Authorized interface {
	GetAuthorization() string
}

func Authed[In Authorized, Out any](
	api huma.API,
	a *Authenticator,
	op huma.Operation,
	handle func(ctx context.Context, actor *iam.Actor, in *In) (Out, error),
) {
	huma.Register(api, op, func(ctx context.Context, in *In) (*httpx.Response[Out], error) {
		actor, _, err := a.Authenticate(ctx, (*in).GetAuthorization())
		if err != nil {
			return nil, httpx.ErrorFrom(ctx, err)
		}
		out, err := handle(ctx, actor, in)
		return httpx.Return(ctx, out, err)
	})
}

func Bearer[In Authorized, Out any](
	api huma.API,
	a *Authenticator,
	op huma.Operation,
	handle func(ctx context.Context, token string, in *In) (Out, error),
) {
	huma.Register(api, op, func(ctx context.Context, in *In) (*httpx.Response[Out], error) {
		token, err := a.ExtractToken((*in).GetAuthorization())
		if err != nil {
			return nil, httpx.ErrorFrom(ctx, err)
		}
		out, err := handle(ctx, token, in)
		return httpx.Return(ctx, out, err)
	})
}
