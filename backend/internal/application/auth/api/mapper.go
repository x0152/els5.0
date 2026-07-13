package api

import (
	"fmt"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

func toLoginStartCommand(in *LoginStartInput) usecases.LoginStartCommand {
	return usecases.LoginStartCommand{
		Email:    in.Body.Email,
		Password: in.Body.Password,
	}
}

func toLoginConfirmCommand(in *LoginConfirmInput) usecases.LoginConfirmCommand {
	return usecases.LoginConfirmCommand{Token: in.Body.Token}
}

func toSetPasswordCommand(in *SetPasswordInput) usecases.SetPasswordCommand {
	return usecases.SetPasswordCommand{
		Token:           in.Body.Token,
		Password:        in.Body.Password,
		PasswordConfirm: in.Body.PasswordConfirm,
	}
}

func toResendInviteCommand(in *ResendInviteInput) usecases.ResendInviteCommand {
	return usecases.ResendInviteCommand{Email: in.Body.Email}
}

func toForgotPasswordCommand(in *ForgotPasswordInput) usecases.ForgotPasswordCommand {
	return usecases.ForgotPasswordCommand{Email: in.Body.Email}
}

func toResetPasswordCommand(in *ResetPasswordInput) usecases.ResetPasswordCommand {
	return usecases.ResetPasswordCommand{
		Token:           in.Body.Token,
		Password:        in.Body.Password,
		PasswordConfirm: in.Body.PasswordConfirm,
	}
}

func toImpersonateCommand(in *ImpersonateInput) (usecases.ImpersonateCommand, error) {
	id, err := parseAccountID(in.Body.AccountID)
	if err != nil {
		return usecases.ImpersonateCommand{}, err
	}
	return usecases.ImpersonateCommand{TargetAccountID: id}, nil
}

func parseAccountID(s string) (iam.AccountID, error) {
	id, err := vo.ParseID(s)
	if err != nil {
		return iam.AccountID{}, fmt.Errorf("%w: account_id: %v", shared.ErrValidation, err)
	}
	return iam.AccountID{ID: id}, nil
}

func toSessionOutput(r usecases.LoginResult) SessionOutput {
	return SessionOutput{
		AccountID: r.Account.ID().String(),
		Email:     r.Account.Email().String(),
		Token:     r.Token,
		ExpiresAt: r.ExpiresAt,
	}
}

func toResendInviteOutput(r usecases.ResendInviteResult) ResendInviteOutput {
	var out ResendInviteOutput
	out.Body.SentTo = r.SentTo
	return out
}

func toForgotPasswordOutput(r usecases.ForgotPasswordResult) ForgotPasswordOutput {
	var out ForgotPasswordOutput
	out.Body.SentTo = r.SentTo
	return out
}

func toSetPasswordOutput(r usecases.SetPasswordResult) PasswordChangedOutput {
	return PasswordChangedOutput{Email: r.Email}
}

func toResetPasswordOutput(r usecases.ResetPasswordResult) PasswordChangedOutput {
	return PasswordChangedOutput{Email: r.Email}
}
