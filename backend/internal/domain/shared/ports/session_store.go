package ports

import (
	"context"
	"time"
)

type SessionSubject struct {
	AccountID     string
	Email         string
	Role          string
	EntityID      string
	IsGlobalAdmin bool
}

func (s SessionSubject) IsZero() bool { return s.AccountID == "" }

type SessionStore interface {
	Create(ctx context.Context, subject SessionSubject, ttl time.Duration) (token string, err error)
	Lookup(ctx context.Context, token string) (SessionSubject, error)
	Revoke(ctx context.Context, token string) error
	RevokeByAccountID(ctx context.Context, accountID string) error
}
