package ports

import (
	"context"
	"time"
)

type InviteTokenPurpose string

const (
	InviteTokenSetPassword   InviteTokenPurpose = "set_password"
	InviteTokenMagicLogin    InviteTokenPurpose = "magic_login"
	InviteTokenResetPassword InviteTokenPurpose = "reset_password"
)

type InviteToken struct {
	Purpose   InviteTokenPurpose
	AccountID string
	Reusable  bool
}

type InviteTokenStore interface {
	Issue(ctx context.Context, tok InviteToken, ttl time.Duration) (token string, err error)
	Consume(ctx context.Context, token string) (InviteToken, error)
}
