package agent

import (
	"context"
	"time"
)

type Session struct {
	ID               string
	AccountID        string
	Model            string
	ContextStartedAt time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Repository interface {
	GetOrCreateSession(ctx context.Context, accountID string) (Session, error)
	UpdateModel(ctx context.Context, sessionID, model string) error
	ResetContext(ctx context.Context, sessionID string, at time.Time) error
	DeleteMessages(ctx context.Context, sessionID string) error
	DeleteMessagesFrom(ctx context.Context, sessionID string, from time.Time) error
	InsertMessage(ctx context.Context, m Message) error
	ListMessages(ctx context.Context, sessionID string) ([]Message, error)
	ListMessagesSince(ctx context.Context, sessionID string, since time.Time) ([]Message, error)
}
