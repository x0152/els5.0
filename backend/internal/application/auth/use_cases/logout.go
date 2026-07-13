package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/shared/ports"
)

type LogoutUseCase struct {
	sessions ports.SessionStore
}

func NewLogoutUseCase(sessions ports.SessionStore) *LogoutUseCase {
	return &LogoutUseCase{sessions: sessions}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, token string) error {
	// 1. Revoke the session in the store; Revoke is idempotent.
	return uc.sessions.Revoke(ctx, token)
}
