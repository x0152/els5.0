package usecases

import (
	"context"
	"errors"

	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/timex"
)

type TodayResult struct {
	Question string
	Entry    *diary.Entry
	Warmup   []diary.Correction
	Streak   int
}

type GetTodayUseCase struct {
	repo   diary.Repository
	worker *ReplyWorker
	clock  timex.Clock
}

func NewGetTodayUseCase(repo diary.Repository, worker *ReplyWorker, clock timex.Clock) *GetTodayUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &GetTodayUseCase{repo: repo, worker: worker, clock: clock}
}

func (uc *GetTodayUseCase) Execute(ctx context.Context, actor *iam.Actor) (TodayResult, error) {
	// 1. Resolve "today" in the platform timezone.
	accountID := actor.AccountID().String()
	now := uc.clock.Now().In(timex.MSK)
	today := timex.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0)

	// 2. Load today's entry if already written.
	var entry *diary.Entry
	if e, err := uc.repo.GetByDate(ctx, accountID, today); err == nil {
		entry = &e
		// Retry a reply that failed or was lost on a server restart.
		if e.Status == diary.StatusPending && uc.worker != nil {
			uc.worker.Kick(e, actor.Account().NativeLanguage())
		}
	} else if !errors.Is(err, shared.ErrNotFound) {
		return TodayResult{}, err
	}

	// 3. Take the question and warmup from the latest previous entry.
	question := diary.DefaultQuestion(now)
	var warmup []diary.Correction
	latest, err := uc.repo.Latest(ctx, accountID, 2)
	if err != nil {
		return TodayResult{}, err
	}
	for _, prev := range latest {
		if diary.SameDay(prev.Date, today) {
			continue
		}
		if prev.NextQuestion != "" {
			question = prev.NextQuestion
		}
		warmup = prev.Corrections
		break
	}
	if entry != nil && entry.Question != "" {
		question = entry.Question
	}

	// 4. Count the streak.
	dates, err := uc.repo.Dates(ctx, accountID, 365)
	if err != nil {
		return TodayResult{}, err
	}
	streak := diary.Streak(dates, now)

	return TodayResult{Question: question, Entry: entry, Warmup: warmup, Streak: streak}, nil
}
