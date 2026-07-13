package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/httpx"
)

type Deps struct {
	Authenticator  *authx.Authenticator
	LoginStart     *usecases.LoginStartUseCase
	LoginConfirm   *usecases.LoginConfirmUseCase
	SetPassword    *usecases.SetPasswordUseCase
	ResendInvite   *usecases.ResendInviteUseCase
	ForgotPassword *usecases.ForgotPasswordUseCase
	ResetPassword  *usecases.ResetPasswordUseCase
	Logout         *usecases.LogoutUseCase
	Impersonate    *usecases.ImpersonateUseCase
}

func Register(api huma.API, deps Deps) {
	httpx.Handle(api, huma.Operation{
		OperationID: "loginStart",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Sign in with email and password",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *LoginStartInput) (SessionOutput, error) {
		res, err := deps.LoginStart.Execute(ctx, toLoginStartCommand(in))
		if err != nil {
			return SessionOutput{}, err
		}
		return toSessionOutput(res), nil
	})

	httpx.Handle(api, huma.Operation{
		OperationID: "loginConfirm",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login/confirm",
		Summary:     "Confirm magic login and issue a session token",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *LoginConfirmInput) (SessionOutput, error) {
		res, err := deps.LoginConfirm.Execute(ctx, toLoginConfirmCommand(in))
		if err != nil {
			return SessionOutput{}, err
		}
		return toSessionOutput(res), nil
	})

	httpx.Handle(api, huma.Operation{
		OperationID: "setPassword",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/set-password",
		Summary:     "Set initial password via invite token (single-use); user must log in afterwards",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *SetPasswordInput) (PasswordChangedOutput, error) {
		res, err := deps.SetPassword.Execute(ctx, toSetPasswordCommand(in))
		if err != nil {
			return PasswordChangedOutput{}, err
		}
		return toSetPasswordOutput(res), nil
	})

	httpx.Handle(api, huma.Operation{
		OperationID:   "resendInvite",
		Method:        http.MethodPost,
		Path:          "/api/v1/auth/resend-invite",
		Summary:       "Resend set-password invite to a pending account",
		Tags:          []string{"auth"},
		DefaultStatus: http.StatusAccepted,
	}, func(ctx context.Context, in *ResendInviteInput) (ResendInviteOutput, error) {
		res, err := deps.ResendInvite.Execute(ctx, toResendInviteCommand(in))
		if err != nil {
			return ResendInviteOutput{}, err
		}
		return toResendInviteOutput(res), nil
	})

	httpx.Handle(api, huma.Operation{
		OperationID:   "forgotPassword",
		Method:        http.MethodPost,
		Path:          "/api/v1/auth/forgot-password",
		Summary:       "Request password reset email",
		Tags:          []string{"auth"},
		DefaultStatus: http.StatusAccepted,
	}, func(ctx context.Context, in *ForgotPasswordInput) (ForgotPasswordOutput, error) {
		res, err := deps.ForgotPassword.Execute(ctx, toForgotPasswordCommand(in))
		if err != nil {
			return ForgotPasswordOutput{}, err
		}
		return toForgotPasswordOutput(res), nil
	})

	httpx.Handle(api, huma.Operation{
		OperationID: "resetPassword",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/reset-password",
		Summary:     "Reset password via reset-password token (single-use, 1h TTL); user must log in afterwards",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *ResetPasswordInput) (PasswordChangedOutput, error) {
		res, err := deps.ResetPassword.Execute(ctx, toResetPasswordCommand(in))
		if err != nil {
			return PasswordChangedOutput{}, err
		}
		return toResetPasswordOutput(res), nil
	})

	authx.Bearer(api, deps.Authenticator, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/logout",
		Summary:     "Revoke current session",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, token string, _ *LogoutInput) (EmptyOutput, error) {
		return EmptyOutput{}, deps.Logout.Execute(ctx, token)
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "impersonate",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/impersonate",
		Summary:     "Issue a session token for another account (dev-only, global admin)",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, actor *iam.Actor, in *ImpersonateInput) (SessionOutput, error) {
		cmd, err := toImpersonateCommand(in)
		if err != nil {
			return SessionOutput{}, err
		}
		res, err := deps.Impersonate.Execute(ctx, actor, cmd)
		if err != nil {
			return SessionOutput{}, err
		}
		return toSessionOutput(res), nil
	})
}
