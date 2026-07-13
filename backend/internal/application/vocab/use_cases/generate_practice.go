package usecases

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

const (
	practicePoolSize = 8
	practiceMinWords = 4
)

// PracticeGenerator runs practice generation in the background.
type PracticeGenerator interface {
	Enqueue(accountID, sessionID string, units []vocab.Unit)
}

type GeneratePracticeUseCase struct {
	units     vocab.Repository
	sessions  vocab.PracticeSessionRepository
	generator PracticeGenerator
}

func NewGeneratePracticeUseCase(units vocab.Repository, sessions vocab.PracticeSessionRepository, generator PracticeGenerator) *GeneratePracticeUseCase {
	return &GeneratePracticeUseCase{units: units, sessions: sessions, generator: generator}
}

func (uc *GeneratePracticeUseCase) Execute(ctx context.Context, actor *iam.Actor) (vocab.PracticeSession, error) {
	accountID := actor.AccountID().String()

	// 1. Take words in learning status.
	units, _, err := uc.units.List(ctx, accountID, vocab.ListFilter{Status: vocab.StatusLearning, Limit: 100})
	if err != nil {
		return vocab.PracticeSession{}, err
	}
	if len(units) < practiceMinWords {
		return vocab.PracticeSession{}, shared.Validation(fmt.Errorf("words: need at least %d learning words to practice", practiceMinWords))
	}

	// 2. Random sample up to practicePoolSize words per session.
	rand.Shuffle(len(units), func(i, j int) { units[i], units[j] = units[j], units[i] })
	if len(units) > practicePoolSize {
		units = units[:practicePoolSize]
	}

	// 3. Save the session as generating; generation itself runs in the background (read via polling).
	session := vocab.PracticeSession{
		ID:      uuid.NewString(),
		Words:   units,
		Status:  vocab.PracticeStatusGenerating,
		Answers: map[string]vocab.PracticeAnswer{},
	}
	if err := uc.sessions.Create(ctx, accountID, session); err != nil {
		return vocab.PracticeSession{}, err
	}
	uc.generator.Enqueue(accountID, session.ID, units)
	return session, nil
}
