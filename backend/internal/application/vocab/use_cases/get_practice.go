package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/vocab"
)

type GetPracticeUseCase struct {
	sessions vocab.PracticeSessionRepository
}

func NewGetPracticeUseCase(sessions vocab.PracticeSessionRepository) *GetPracticeUseCase {
	return &GetPracticeUseCase{sessions: sessions}
}

func (uc *GetPracticeUseCase) Execute(ctx context.Context, actor *iam.Actor) (vocab.PracticeSession, error) {
	return uc.sessions.Load(ctx, actor.AccountID().String())
}

type SavePracticeProgressUseCase struct {
	sessions vocab.PracticeSessionRepository
}

func NewSavePracticeProgressUseCase(sessions vocab.PracticeSessionRepository) *SavePracticeProgressUseCase {
	return &SavePracticeProgressUseCase{sessions: sessions}
}

func (uc *SavePracticeProgressUseCase) Execute(ctx context.Context, actor *iam.Actor, sessionID string, answers map[string]vocab.PracticeAnswer, completed bool) error {
	return uc.sessions.SaveProgress(ctx, actor.AccountID().String(), sessionID, answers, completed)
}
