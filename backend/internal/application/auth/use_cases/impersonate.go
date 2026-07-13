package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/utils/timex"
)

type ImpersonateUseCase struct {
	enabled    bool
	accounts   iam.AccountRepository
	roles      iam.AccountRoleRepository
	sessions   ports.SessionStore
	sessionTTL time.Duration
	clock      timex.Clock
}

func NewImpersonateUseCase(
	enabled bool,
	accounts iam.AccountRepository,
	roles iam.AccountRoleRepository,
	sessions ports.SessionStore,
	sessionTTL time.Duration,
	clock timex.Clock,
) *ImpersonateUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &ImpersonateUseCase{
		enabled:    enabled,
		accounts:   accounts,
		roles:      roles,
		sessions:   sessions,
		sessionTTL: sessionTTL,
		clock:      clock,
	}
}

type ImpersonateCommand struct {
	TargetAccountID iam.AccountID
}

func (uc *ImpersonateUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd ImpersonateCommand) (LoginResult, error) {
	// 1. Feature is disabled outside the dev environment — deny unconditionally.
	if !uc.enabled {
		return LoginResult{}, fmt.Errorf("%w: impersonation is disabled", shared.ErrForbidden)
	}

	// 2. Only a global administrator may impersonate.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return LoginResult{}, err
	}

	// 3. Do not allow impersonating yourself.
	if actor.AccountID() == cmd.TargetAccountID {
		return LoginResult{}, fmt.Errorf("%w: cannot impersonate self", shared.ErrValidation)
	}

	// 4. Load the target account and verify it is active.
	target, err := loadActiveAccount(ctx, uc.accounts, cmd.TargetAccountID.String())
	if err != nil {
		if errors.Is(err, shared.ErrUnauthorized) {
			return LoginResult{}, fmt.Errorf("%w: target account not found", shared.ErrNotFound)
		}
		return LoginResult{}, err
	}

	// 5. Issue a new session for the target account (same as a normal login).
	return issueSession(ctx, uc.sessions, uc.roles, target, uc.sessionTTL, uc.clock)
}
