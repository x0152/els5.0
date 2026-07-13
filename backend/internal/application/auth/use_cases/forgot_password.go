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

type ForgotPasswordUseCase struct {
	accounts iam.AccountRepository
	invites  ports.InviteTokenStore
	mail     ports.MailSender
	ttl      time.Duration
	linkTmpl string
}

type ForgotPasswordDeps struct {
	Accounts iam.AccountRepository
	Invites  ports.InviteTokenStore
	Mail     ports.MailSender
	TTL      time.Duration
	LinkTmpl string
}

func NewForgotPasswordUseCase(d ForgotPasswordDeps) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		accounts: d.Accounts,
		invites:  d.Invites,
		mail:     d.Mail,
		ttl:      d.TTL,
		linkTmpl: d.LinkTmpl,
	}
}

type ForgotPasswordCommand struct {
	Email string
}

type ForgotPasswordResult struct {
	SentTo string
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, cmd ForgotPasswordCommand) (ForgotPasswordResult, error) {
	// 1. Build the mask ahead of time — return it on any failure to avoid revealing account existence.
	result := ForgotPasswordResult{SentTo: maskEmail(cmd.Email)}

	// 2. Parse the email; bad format → quietly return the mask.
	email, err := vo.NewEmail(cmd.Email)
	if err != nil {
		return result, nil
	}

	// 3. Look up the account; missing or blocked status is hidden behind the mask.
	acc, err := uc.accounts.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return result, nil
		}
		return ForgotPasswordResult{}, err
	}
	if acc.Status() == iam.AccountStatusBlocked || acc.Status() == iam.AccountStatusNoAuth {
		return result, nil
	}

	// 4. Issue a reset token.
	token, err := uc.invites.Issue(ctx, ports.InviteToken{
		Purpose:   ports.InviteTokenResetPassword,
		AccountID: acc.ID().String(),
	}, uc.ttl)
	if err != nil {
		return ForgotPasswordResult{}, err
	}

	// 5. Send the email with the link.
	if err := uc.mail.SendPasswordReset(ctx, acc.Email().String(), acc.Name().Full(), renderLink(uc.linkTmpl, token)); err != nil {
		return ForgotPasswordResult{}, err
	}
	return ForgotPasswordResult{SentTo: maskEmail(acc.Email().String())}, nil
}
