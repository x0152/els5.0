package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/domain/shared/vo"
)

type ResendInviteUseCase struct {
	accounts iam.AccountRepository
	invites  ports.InviteTokenStore
	mail     ports.MailSender
	ttl      time.Duration
	linkTmpl string
}

type ResendInviteDeps struct {
	Accounts iam.AccountRepository
	Invites  ports.InviteTokenStore
	Mail     ports.MailSender
	TTL      time.Duration
	LinkTmpl string
}

func NewResendInviteUseCase(d ResendInviteDeps) *ResendInviteUseCase {
	return &ResendInviteUseCase{
		accounts: d.Accounts,
		invites:  d.Invites,
		mail:     d.Mail,
		ttl:      d.TTL,
		linkTmpl: d.LinkTmpl,
	}
}

type ResendInviteCommand struct {
	Email string
}

type ResendInviteResult struct {
	SentTo string
}

func (uc *ResendInviteUseCase) Execute(ctx context.Context, cmd ResendInviteCommand) (ResendInviteResult, error) {
	// 1. The mask is returned on any failure to avoid revealing account existence.
	result := ResendInviteResult{SentTo: maskEmail(cmd.Email)}

	// 2. Parse the email; bad format → quietly return the mask.
	email, err := vo.NewEmail(cmd.Email)
	if err != nil {
		return result, nil
	}

	// 3. Look up the account; resend works only for pending_password, otherwise the mask.
	acc, err := uc.accounts.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return result, nil
		}
		return ResendInviteResult{}, err
	}
	if acc.Status() != iam.AccountStatusPendingPassword {
		return result, nil
	}

	// 4. Re-issue the invite token.
	token, err := uc.invites.Issue(ctx, ports.InviteToken{
		Purpose:   ports.InviteTokenSetPassword,
		AccountID: acc.ID().String(),
	}, uc.ttl)
	if err != nil {
		return ResendInviteResult{}, err
	}

	// 5. Send the email.
	if err := uc.mail.SendSetPasswordInvite(ctx, acc.Email().String(), acc.Name().Full(), renderLink(uc.linkTmpl, token)); err != nil {
		return ResendInviteResult{}, err
	}
	return ResendInviteResult{SentTo: maskEmail(acc.Email().String())}, nil
}
