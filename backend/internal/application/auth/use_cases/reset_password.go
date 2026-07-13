package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
)

type ResetPasswordUseCase struct {
	accounts    iam.AccountRepository
	credentials iam.CredentialsRepository
	hasher      ports.PasswordHasher
	invites     ports.InviteTokenStore
	sessions    ports.SessionStore
	policy      iam.PasswordPolicy
}

type ResetPasswordDeps struct {
	Accounts    iam.AccountRepository
	Credentials iam.CredentialsRepository
	Hasher      ports.PasswordHasher
	Invites     ports.InviteTokenStore
	Sessions    ports.SessionStore
	Policy      iam.PasswordPolicy
}

func NewResetPasswordUseCase(d ResetPasswordDeps) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		accounts:    d.Accounts,
		credentials: d.Credentials,
		hasher:      d.Hasher,
		invites:     d.Invites,
		sessions:    d.Sessions,
		policy:      d.Policy,
	}
}

type ResetPasswordCommand struct {
	Token           string
	Password        string
	PasswordConfirm string
}

type ResetPasswordResult struct {
	Email string
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, cmd ResetPasswordCommand) (ResetPasswordResult, error) {
	// 1. Validate the password policy and confirm match.
	if err := uc.policy.Compare(cmd.Password, cmd.PasswordConfirm); err != nil {
		return ResetPasswordResult{}, err
	}

	// 2. Consume the one-time reset token (TTL/single-use is guaranteed by InviteTokenStore).
	invite, err := uc.invites.Consume(ctx, cmd.Token)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	if invite.Purpose != ports.InviteTokenResetPassword {
		return ResetPasswordResult{}, fmt.Errorf("%w: wrong token purpose", shared.ErrUnauthorized)
	}

	// 3. Load the account; a blocked account cannot be reset.
	id, err := iamID(invite.AccountID)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	acc, err := uc.accounts.GetByID(ctx, id)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	if acc.Status() == iam.AccountStatusBlocked {
		return ResetPasswordResult{}, fmt.Errorf("%w: account is blocked", shared.ErrForbidden)
	}

	// 4. Hash the new password and save credentials.
	hash, err := uc.hasher.Hash(cmd.Password)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	cred, err := iam.NewCredentials(acc.ID(), hash)
	if err != nil {
		return ResetPasswordResult{}, err
	}
	if err := uc.credentials.Save(ctx, cred); err != nil {
		return ResetPasswordResult{}, err
	}

	// 5. If the account was pending_password — activate it after password setup.
	if acc.Status() == iam.AccountStatusPendingPassword {
		if err := acc.Activate(); err != nil {
			return ResetPasswordResult{}, err
		}
		if err := uc.accounts.Update(ctx, acc); err != nil {
			return ResetPasswordResult{}, err
		}
	}

	// 6. Any previously issued sessions become invalid after a password change.
	if uc.sessions != nil {
		if err := uc.sessions.RevokeByAccountID(ctx, acc.ID().String()); err != nil {
			return ResetPasswordResult{}, err
		}
	}

	// 7. Do NOT issue a session: the user must sign in via /login as usual.
	return ResetPasswordResult{Email: acc.Email().String()}, nil
}
