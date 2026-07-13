package usecases

import (
	"errors"
	"fmt"
	"time"

	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/timex"
)

type LoginConfirmUseCase struct {
	accounts   iam.AccountRepository
	roles      iam.AccountRoleRepository
	invites    ports.InviteTokenStore
	sessions   ports.SessionStore
	sessionTTL time.Duration
	clock      timex.Clock
}

func NewLoginConfirmUseCase(
	accounts iam.AccountRepository,
	roles iam.AccountRoleRepository,
	invites ports.InviteTokenStore,
	sessions ports.SessionStore,
	sessionTTL time.Duration,
	clock timex.Clock,
) *LoginConfirmUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &LoginConfirmUseCase{
		accounts:   accounts,
		roles:      roles,
		invites:    invites,
		sessions:   sessions,
		sessionTTL: sessionTTL,
		clock:      clock,
	}
}

type LoginConfirmCommand struct {
	Token string
}

type LoginResult struct {
	Account   *iam.Account
	Token     string
	ExpiresAt time.Time
}

func (uc *LoginConfirmUseCase) Execute(ctx context.Context, cmd LoginConfirmCommand) (LoginResult, error) {
	// 1. Consume the one-time token.
	invite, err := uc.invites.Consume(ctx, cmd.Token)
	if err != nil {
		return LoginResult{}, err
	}
	if invite.Purpose != ports.InviteTokenMagicLogin {
		return LoginResult{}, fmt.Errorf("%w: wrong token purpose", shared.ErrUnauthorized)
	}

	// 2. Load the account and verify it is active.
	acc, err := loadActiveAccount(ctx, uc.accounts, invite.AccountID)
	if err != nil {
		return LoginResult{}, err
	}

	// 3. Issue a session token.
	return issueSession(ctx, uc.sessions, uc.roles, acc, uc.sessionTTL, uc.clock)
}

func loadActiveAccount(ctx context.Context, accounts iam.AccountRepository, accountID string) (*iam.Account, error) {
	id, err := vo.ParseID(accountID)
	if err != nil {
		return nil, shared.ErrUnauthorized
	}
	acc, err := accounts.GetByID(ctx, iam.AccountID{ID: id})
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.ErrUnauthorized
		}
		return nil, err
	}
	if err := acc.EnsureCanLogin(); err != nil {
		return nil, err
	}
	return acc, nil
}

func issueSession(
	ctx context.Context,
	sessions ports.SessionStore,
	roles iam.AccountRoleRepository,
	acc *iam.Account,
	ttl time.Duration,
	clock timex.Clock,
) (LoginResult, error) {
	link, err := roles.GetByAccountID(ctx, acc.ID())
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return LoginResult{}, fmt.Errorf("%w: account has no role", shared.ErrForbidden)
		}
		return LoginResult{}, fmt.Errorf("resolve role: %w", err)
	}
	subject := ports.SessionSubject{
		AccountID:     acc.ID().String(),
		Email:         acc.Email().String(),
		Role:          link.Role.String(),
		EntityID:      link.EntityID.String(),
		IsGlobalAdmin: link.Role == iam.RoleAdmin,
	}
	token, err := sessions.Create(ctx, subject, ttl)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Account: acc, Token: token, ExpiresAt: clock.Now().Add(ttl)}, nil
}
