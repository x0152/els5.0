package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/utils/timex"
)

type StartLessonResult struct {
	Lesson     *workout.Lesson
	Generating bool
}

type StartLessonUseCase struct {
	repo     workout.Repository
	generate *GenerateLessonUseCase
	llm      LLMClient
	logger   *slog.Logger
	clock    timex.Clock
}

func NewStartLessonUseCase(repo workout.Repository, generate *GenerateLessonUseCase, llm LLMClient, logger *slog.Logger, clock timex.Clock) *StartLessonUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &StartLessonUseCase{repo: repo, generate: generate, llm: llm, logger: logger, clock: clock}
}

func (uc *StartLessonUseCase) Execute(ctx context.Context, accountID string) (StartLessonResult, error) {
	// 1. An active lesson is returned right away.
	if lesson, err := uc.repo.CurrentLesson(ctx, accountID); err == nil {
		return StartLessonResult{Lesson: &lesson}, nil
	} else if !errors.Is(err, shared.ErrNotFound) {
		return StartLessonResult{}, err
	}
	if !uc.llm.Available() {
		return StartLessonResult{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. A live generation claim in the DB means work is already in progress.
	now := uc.clock.Now().In(timex.MSK)
	pending, err := uc.repo.PendingLesson(ctx, accountID)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return StartLessonResult{}, err
	}
	if err == nil && pending.GenerationInFlight(now) {
		return StartLessonResult{Generating: true}, nil
	}

	// 3. Claim the slot synchronously so the UI flips to "generating" right away;
	// a concurrent claim surfaces as a conflict and is treated as already running.
	recent, err := uc.repo.ListRecentLessons(ctx, accountID, 1)
	if err != nil {
		return StartLessonResult{}, err
	}
	number := 1
	if len(recent) > 0 {
		number = recent[0].Number + 1
	}
	claim := workout.Lesson{ID: uuid.NewString(), AccountID: accountID, Number: number,
		Status: workout.LessonStatusGenerating, Steps: []workout.Step{}, CreatedAt: now}
	if err := uc.repo.ClaimGeneration(ctx, claim, now.Add(-workout.GenerationStaleAfter)); err != nil {
		if errors.Is(err, shared.ErrConflict) {
			return StartLessonResult{Generating: true}, nil
		}
		return StartLessonResult{}, err
	}

	// 4. Generate into the claim in the background.
	go func() {
		genCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Minute)
		defer cancel()
		if _, err := uc.generate.Execute(genCtx, accountID); err != nil && !errors.Is(err, shared.ErrConflict) {
			uc.logger.Error("workout lesson generation failed", slog.String("account", accountID), slog.String("err", err.Error()))
		}
	}()
	return StartLessonResult{Generating: true}, nil
}
