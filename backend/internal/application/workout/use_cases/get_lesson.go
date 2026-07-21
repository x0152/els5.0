package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/workout"
)

type GetLessonUseCase struct {
	repo workout.Repository
}

func NewGetLessonUseCase(repo workout.Repository) *GetLessonUseCase {
	return &GetLessonUseCase{repo: repo}
}

func (uc *GetLessonUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) (workout.Lesson, error) {
	// 1. Lessons are private to their owner.
	return uc.repo.GetLesson(ctx, actor.AccountID().String(), id)
}
