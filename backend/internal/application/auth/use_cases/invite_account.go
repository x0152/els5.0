package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
)

type InviteAccountUseCase struct {
	accounts iam.AccountRepository
	invites  ports.InviteTokenStore
	mail     ports.MailSender
	ttl      time.Duration
	linkTmpl string
}

type InviteAccountDeps struct {
	Accounts iam.AccountRepository
	Invites  ports.InviteTokenStore
	Mail     ports.MailSender
	TTL      time.Duration
	LinkTmpl string
}

func NewInviteAccountUseCase(d InviteAccountDeps) *InviteAccountUseCase {
	return &InviteAccountUseCase{
		accounts: d.Accounts,
		invites:  d.Invites,
		mail:     d.Mail,
		ttl:      d.TTL,
		linkTmpl: d.LinkTmpl,
	}
}

type InviteAccountCommand struct {
	Email     string
	FirstName string
	LastName  string
}

func (uc *InviteAccountUseCase) Execute(ctx context.Context, cmd InviteAccountCommand) (*iam.Account, error) {
	// 1. Build a domain Account in pending_password status.
	acc, err := iam.NewPendingAccountNow(iam.NewAccountID(), cmd.Email, cmd.FirstName, cmd.LastName)
	if err != nil {
		return nil, err
	}

	// 2. Persist the account (email uniqueness is enforced in the DB).
	if err := uc.accounts.Create(ctx, acc); err != nil {
		return nil, err
	}

	// 3. Issue a one-time invite token for password setup.
	token, err := uc.invites.Issue(ctx, ports.InviteToken{
		Purpose:   ports.InviteTokenSetPassword,
		AccountID: acc.ID().String(),
	}, uc.ttl)
	if err != nil {
		return nil, err
	}

	// 4. Send the email (stub → stdout).
	if err := uc.mail.SendSetPasswordInvite(ctx, acc.Email().String(), acc.Name().Full(), renderLink(uc.linkTmpl, token)); err != nil {
		return nil, err
	}

	return acc, nil
}

func (uc *InviteAccountUseCase) ResendFor(ctx context.Context, acc *iam.Account) error {
	if acc.Status() != iam.AccountStatusPendingPassword {
		return fmt.Errorf("%w: account is not pending password", shared.ErrConflict)
	}
	token, err := uc.invites.Issue(ctx, ports.InviteToken{
		Purpose:   ports.InviteTokenSetPassword,
		AccountID: acc.ID().String(),
	}, uc.ttl)
	if err != nil {
		return err
	}
	return uc.mail.SendSetPasswordInvite(ctx, acc.Email().String(), acc.Name().Full(), renderLink(uc.linkTmpl, token))
}
