package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/timex"
)

type LoginStartUseCase struct {
	accounts        iam.AccountRepository
	credentials     iam.CredentialsRepository
	roles           iam.AccountRoleRepository
	hasher          ports.PasswordHasher
	sessions        ports.SessionStore
	attempts        ports.LoginAttemptStore
	lockoutAttempts int
	lockoutWindow   time.Duration
	sessionTTL      time.Duration
	clock           timex.Clock
}

type LoginStartDeps struct {
	Accounts        iam.AccountRepository
	Credentials     iam.CredentialsRepository
	Roles           iam.AccountRoleRepository
	Hasher          ports.PasswordHasher
	Sessions        ports.SessionStore
	Attempts        ports.LoginAttemptStore
	LockoutAttempts int
	LockoutWindow   time.Duration
	SessionTTL      time.Duration
	Clock           timex.Clock
}

func NewLoginStartUseCase(d LoginStartDeps) *LoginStartUseCase {
	clock := d.Clock
	if clock == nil {
		clock = timex.System()
	}
	return &LoginStartUseCase{
		accounts:        d.Accounts,
		credentials:     d.Credentials,
		roles:           d.Roles,
		hasher:          d.Hasher,
		sessions:        d.Sessions,
		attempts:        d.Attempts,
		lockoutAttempts: d.LockoutAttempts,
		lockoutWindow:   d.LockoutWindow,
		sessionTTL:      d.SessionTTL,
		clock:           clock,
	}
}

type LoginStartCommand struct {
	Email    string
	Password string
}

func (uc *LoginStartUseCase) Execute(ctx context.Context, cmd LoginStartCommand) (LoginResult, error) {
	deny := fmt.Errorf("%w: invalid email or password", shared.ErrUnauthorized)

	// 1. Parse the email.
	email, err := vo.NewEmail(cmd.Email)
	if err != nil {
		return LoginResult{}, deny
	}

	// 2. Look up the account; missing or non-active → same unauthorized response.
	acc, err := uc.accounts.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return LoginResult{}, deny
		}
		return LoginResult{}, err
	}
	if err := acc.EnsureCanLogin(); err != nil {
		return LoginResult{}, deny
	}

	// 3. Brute-force lockout.
	if uc.attempts != nil {
		locked, err := uc.attempts.IsLocked(ctx, acc.ID().String())
		if err != nil {
			return LoginResult{}, err
		}
		if locked {
			return LoginResult{}, deny
		}
	}

	// 4. Verify password.
	cred, err := uc.credentials.GetByAccountID(ctx, acc.ID())
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return LoginResult{}, deny
		}
		return LoginResult{}, err
	}
	if err := cred.Verify(cmd.Password, uc.hasher); err != nil {
		if uc.attempts != nil {
			if ferr := uc.attempts.Fail(ctx, acc.ID().String(), uc.lockoutAttempts, uc.lockoutWindow); ferr != nil {
				return LoginResult{}, ferr
			}
		}
		return LoginResult{}, deny
	}
	if uc.attempts != nil {
		if err := uc.attempts.Reset(ctx, acc.ID().String()); err != nil {
			return LoginResult{}, err
		}
	}

	// 5. Issue a session token.
	return issueSession(ctx, uc.sessions, uc.roles, acc, uc.sessionTTL, uc.clock)
}
