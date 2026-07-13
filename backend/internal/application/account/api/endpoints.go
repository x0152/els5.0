package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/account/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator        *authx.Authenticator
	Me                   *usecases.MeUseCase
	UpdateProfile        *usecases.UpdateProfileUseCase
	ListApps             *usecases.ListAppsUseCase
	UploadAccountPicture *usecases.UploadAccountPictureUseCase
	ImpersonationEnabled bool
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "accountMe",
		Method:      http.MethodGet,
		Path:        "/api/v1/account/me",
		Summary:     "Current account",
		Tags:        []string{"account"},
	}, func(ctx context.Context, actor *iam.Actor, _ *MeInput) (MeOutput, error) {
		res, err := deps.Me.Execute(ctx, actor)
		if err != nil {
			return MeOutput{}, err
		}
		return toMeOutput(res, deps.ImpersonationEnabled), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "accountUpdateProfile",
		Method:      http.MethodPut,
		Path:        "/api/v1/account/me",
		Summary:     "Update current account profile (name, English level, about me)",
		Tags:        []string{"account"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateProfileInput) (MeOutput, error) {
		res, err := deps.UpdateProfile.Execute(ctx, actor, usecases.UpdateProfileCommand{
			FirstName:    in.Body.FirstName,
			LastName:     in.Body.LastName,
			EnglishLevel: in.Body.EnglishLevel,
			AboutMe:      in.Body.AboutMe,
		})
		if err != nil {
			return MeOutput{}, err
		}
		return accountToMeOutput(res.Account, actor), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "accountApps",
		Method:      http.MethodGet,
		Path:        "/api/v1/account/apps",
		Summary:     "List applications available to the current account (for sidebar)",
		Tags:        []string{"account"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListAppsInput) (AppsOutput, error) {
		res, err := deps.ListApps.Execute(ctx, actor)
		if err != nil {
			return AppsOutput{}, err
		}
		return toAppsOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "accountMeUploadPicture",
		Method:      http.MethodPost,
		Path:        "/api/v1/account/me/picture",
		Summary:     "Upload (replace) current account picture",
		Tags:        []string{"account"},
	}, func(ctx context.Context, actor *iam.Actor, in *UploadAccountPictureInput) (MeOutput, error) {
		if deps.UploadAccountPicture == nil {
			return MeOutput{}, huma.Error503ServiceUnavailable("picture upload is not configured")
		}
		form := in.RawBody.Data()
		if form == nil || !form.File.IsSet {
			return MeOutput{}, huma.Error400BadRequest("file is required")
		}
		defer form.File.Close()
		res, err := deps.UploadAccountPicture.Execute(ctx, actor, usecases.UploadAccountPictureCommand{
			Reader:      form.File,
			Size:        form.File.Size,
			ContentType: form.File.ContentType,
			Filename:    form.File.Filename,
		})
		if err != nil {
			return MeOutput{}, err
		}
		return accountToMeOutput(res.Account, actor), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "accountUploadPicture",
		Method:      http.MethodPost,
		Path:        "/api/v1/accounts/{account_id}/picture",
		Summary:     "Upload (replace) picture for any account (global admin only)",
		Tags:        []string{"account"},
	}, func(ctx context.Context, actor *iam.Actor, in *UploadAccountPictureByIDInput) (AccountPictureOutput, error) {
		if deps.UploadAccountPicture == nil {
			return AccountPictureOutput{}, huma.Error503ServiceUnavailable("picture upload is not configured")
		}
		targetID, err := parseAccountID(in.AccountID)
		if err != nil {
			return AccountPictureOutput{}, err
		}
		form := in.RawBody.Data()
		if form == nil || !form.File.IsSet {
			return AccountPictureOutput{}, huma.Error400BadRequest("file is required")
		}
		defer form.File.Close()
		res, err := deps.UploadAccountPicture.Execute(ctx, actor, usecases.UploadAccountPictureCommand{
			TargetID:    targetID,
			Reader:      form.File,
			Size:        form.File.Size,
			ContentType: form.File.ContentType,
			Filename:    form.File.Filename,
		})
		if err != nil {
			return AccountPictureOutput{}, err
		}
		return toAccountPictureOutput(res.Account), nil
	})
}
