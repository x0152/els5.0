package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/admin/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	CreateAdmin   *usecases.CreateAdminUseCase
	GetAdmin      *usecases.GetAdminUseCase
	ListAdmins    *usecases.ListAdminsUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "createAdministrator",
		Method:        http.MethodPost,
		Path:          "/api/v1/administrators",
		Summary:       "Create an administrator (global admin only)",
		Tags:          []string{"administrator"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, actor *iam.Actor, in *CreateAdminInput) (AdminOutput, error) {
		res, err := deps.CreateAdmin.Execute(ctx, actor, toCreateAdminCommand(in))
		if err != nil {
			return AdminOutput{}, err
		}
		return toAdminOutput(res.Admin), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listAdministrators",
		Method:      http.MethodGet,
		Path:        "/api/v1/administrators",
		Summary:     "List administrators (global admin only)",
		Tags:        []string{"administrator"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListAdminsInput) (AdminsOutput, error) {
		res, err := deps.ListAdmins.Execute(ctx, actor, toListAdminsQuery(in))
		if err != nil {
			return AdminsOutput{}, err
		}
		return toAdminsOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getAdministrator",
		Method:      http.MethodGet,
		Path:        "/api/v1/administrators/{id}",
		Summary:     "Get an administrator by id (global admin only)",
		Tags:        []string{"administrator"},
	}, func(ctx context.Context, actor *iam.Actor, in *GetAdminInput) (AdminOutput, error) {
		q, err := toGetAdminQuery(in)
		if err != nil {
			return AdminOutput{}, err
		}
		res, err := deps.GetAdmin.Execute(ctx, actor, q)
		if err != nil {
			return AdminOutput{}, err
		}
		return toAdminOutput(res.Admin), nil
	})
}
