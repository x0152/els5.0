package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/utils/timex"
)

type TodayResult struct {
	Lesson    *workout.Lesson
	Streak    int
	Days      []time.Time
	Completed bool
}

type GetTodayUseCase struct {
	repo  workout.Repository
	clock timex.Clock
}

func NewGetTodayUseCase(repo workout.Repository, clock timex.Clock) *GetTodayUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &GetTodayUseCase{repo: repo, clock: clock}
}

func (uc *GetTodayUseCase) Execute(ctx context.Context, actor *iam.Actor) (TodayResult, error) {
	// 1. The calendar and streak come from completed lesson days.
	accountID := actor.AccountID().String()
	now := uc.clock.Now().In(timex.MSK)
	days, err := uc.repo.ListCompletedDates(ctx, accountID, now.AddDate(0, -4, 0))
	if err != nil {
		return TodayResult{}, err
	}
	result := TodayResult{
		Streak: workout.Streak(days, now),
		Days:   days,
	}
	if len(days) > 0 && days[0].In(timex.MSK).Format("2006-01-02") == now.Format("2006-01-02") {
		result.Completed = true
	}

	// 2. Attach the current lesson when one is waiting.
	lesson, err := uc.repo.CurrentLesson(ctx, accountID)
	if err == nil {
		result.Lesson = &lesson
	} else if !errors.Is(err, shared.ErrNotFound) {
		return TodayResult{}, err
	}
	return result, nil
}
