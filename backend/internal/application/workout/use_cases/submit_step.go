package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/utils/timex"
)

type SubmitStepCommand struct {
	LessonID string
	StepID   string
	Score    int
	Results  []workout.ItemResult
}

type SubmitStepUseCase struct {
	repo  workout.Repository
	clock timex.Clock
}

func NewSubmitStepUseCase(repo workout.Repository, clock timex.Clock) *SubmitStepUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &SubmitStepUseCase{repo: repo, clock: clock}
}

func (uc *SubmitStepUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd SubmitStepCommand) (workout.Lesson, error) {
	// 1. Load the lesson and mark the step done; the entity completes itself on the last step.
	accountID := actor.AccountID().String()
	lesson, err := uc.repo.GetLesson(ctx, accountID, cmd.LessonID)
	if err != nil {
		return workout.Lesson{}, err
	}
	now := uc.clock.Now().In(timex.MSK)
	if err := lesson.SubmitStep(cmd.StepID, cmd.Score, now); err != nil {
		return workout.Lesson{}, err
	}
	if err := uc.repo.UpdateLesson(ctx, lesson); err != nil {
		return workout.Lesson{}, err
	}

	// 2. Feed the per-item results into the spiral material.
	if err := uc.repo.UpsertItems(ctx, accountID, lesson.Number, cmd.Results, now); err != nil {
		return workout.Lesson{}, err
	}
	return lesson, nil
}
