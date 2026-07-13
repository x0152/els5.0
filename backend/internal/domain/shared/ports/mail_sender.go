package ports

import "context"

type MailSender interface {
	SendSetPasswordInvite(ctx context.Context, to, recipientName, link string) error
	SendMagicLogin(ctx context.Context, to, recipientName, link string) error
	SendPasswordReset(ctx context.Context, to, recipientName, link string) error
}
