package usecases

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
)

type SetPasswordUseCase struct {
	accounts    iam.AccountRepository
	credentials iam.CredentialsRepository
	hasher      ports.PasswordHasher
	invites     ports.InviteTokenStore
	sessions    ports.SessionStore
	policy      iam.PasswordPolicy
}

type SetPasswordDeps struct {
	Accounts    iam.AccountRepository
	Credentials iam.CredentialsRepository
	Hasher      ports.PasswordHasher
	Invites     ports.InviteTokenStore
	Sessions    ports.SessionStore
	Policy      iam.PasswordPolicy
}

func NewSetPasswordUseCase(d SetPasswordDeps) *SetPasswordUseCase {
	return &SetPasswordUseCase{
		accounts:    d.Accounts,
		credentials: d.Credentials,
		hasher:      d.Hasher,
		invites:     d.Invites,
		sessions:    d.Sessions,
		policy:      d.Policy,
	}
}

type SetPasswordCommand struct {
	Token           string
	Password        string
	PasswordConfirm string
}

type SetPasswordResult struct {
	Email string
}

func (uc *SetPasswordUseCase) Execute(ctx context.Context, cmd SetPasswordCommand) (SetPasswordResult, error) {
	// 1. Validate the password policy and confirm match.
	if err := uc.policy.Compare(cmd.Password, cmd.PasswordConfirm); err != nil {
		return SetPasswordResult{}, err
	}

	// 2. Consume the one-time invite token (TTL/single-use is guaranteed by InviteTokenStore).
	invite, err := uc.invites.Consume(ctx, cmd.Token)
	if err != nil {
		return SetPasswordResult{}, err
	}
	if invite.Purpose != ports.InviteTokenSetPassword {
		return SetPasswordResult{}, fmt.Errorf("%w: wrong token purpose", shared.ErrUnauthorized)
	}

	// 3. Load the account; a blocked account cannot set a password.
	//    An active account is allowed: the token is one-time/TTL-limited, so
	//    setting a password on top is idempotent (same model as reset) and does not
	//    lead to a stuck 409 if the account was already activated another way.
	id, err := iamID(invite.AccountID)
	if err != nil {
		return SetPasswordResult{}, err
	}
	acc, err := uc.accounts.GetByID(ctx, id)
	if err != nil {
		return SetPasswordResult{}, err
	}
	if acc.Status() == iam.AccountStatusBlocked {
		return SetPasswordResult{}, fmt.Errorf("%w: account is blocked", shared.ErrForbidden)
	}

	// 4. Hash the password and save credentials.
	hash, err := uc.hasher.Hash(cmd.Password)
	if err != nil {
		return SetPasswordResult{}, err
	}
	cred, err := iam.NewCredentials(acc.ID(), hash)
	if err != nil {
		return SetPasswordResult{}, err
	}
	if err := uc.credentials.Save(ctx, cred); err != nil {
		return SetPasswordResult{}, err
	}

	// 5. If the account was pending_password — activate it after password setup.
	if acc.Status() == iam.AccountStatusPendingPassword {
		if err := acc.Activate(); err != nil {
			return SetPasswordResult{}, err
		}
		if err := uc.accounts.Update(ctx, acc); err != nil {
			return SetPasswordResult{}, err
		}
	}

	// 6. Any previously issued sessions become invalid after a password change.
	if uc.sessions != nil {
		if err := uc.sessions.RevokeByAccountID(ctx, acc.ID().String()); err != nil {
			return SetPasswordResult{}, err
		}
	}

	// 7. Do NOT issue a session: the user must sign in via /login as usual.
	return SetPasswordResult{Email: acc.Email().String()}, nil
}
