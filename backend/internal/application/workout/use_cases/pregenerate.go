package usecases

import (
	"context"
	"log/slog"

	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/utils/timex"
)

// PregenerateUseCase keeps the next lesson ready ahead of time: whenever a learner has
// finished their latest lesson, the worker builds the next one in the background.
type PregenerateUseCase struct {
	repo     workout.Repository
	generate *GenerateLessonUseCase
	logger   *slog.Logger
	clock    timex.Clock
}

func NewPregenerateUseCase(repo workout.Repository, generate *GenerateLessonUseCase, logger *slog.Logger) *PregenerateUseCase {
	return &PregenerateUseCase{repo: repo, generate: generate, logger: logger, clock: timex.System()}
}

func (uc *PregenerateUseCase) Execute(ctx context.Context) error {
	// 1. Every learner whose lessons are all completed gets the next one prepared.
	accounts, err := uc.repo.ListAccountsNeedingLesson(ctx, uc.clock.Now().Add(-workout.GenerationStaleAfter))
	if err != nil {
		return err
	}
	for _, accountID := range accounts {
		if _, err := uc.generate.Execute(ctx, accountID); err != nil {
			uc.logger.Error("workout pregeneration failed", slog.String("account", accountID), slog.String("err", err.Error()))
		}
	}
	return nil
}
