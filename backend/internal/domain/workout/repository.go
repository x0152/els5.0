package workout

import (
	"context"
	"time"
)

type Repository interface {
	GetPlan(ctx context.Context, filmID string) (FilmPlan, error)
	SavePlan(ctx context.Context, plan FilmPlan) error
	ListPlannedFilmIDs(ctx context.Context, status string) ([]string, error)
	ListStaleFailedPlanFilmIDs(ctx context.Context, before time.Time) ([]string, error)

	CurrentLesson(ctx context.Context, accountID string) (Lesson, error)
	PendingLesson(ctx context.Context, accountID string) (Lesson, error)
	GetLesson(ctx context.Context, accountID, id string) (Lesson, error)
	ListRecentLessons(ctx context.Context, accountID string, limit int) ([]Lesson, error)
	ClaimGeneration(ctx context.Context, lesson Lesson, staleBefore time.Time) error
	UpdateLesson(ctx context.Context, lesson Lesson) error
	ListCompletedDates(ctx context.Context, accountID string, since time.Time) ([]time.Time, error)
	ListAccountsNeedingLesson(ctx context.Context, staleBefore time.Time) ([]string, error)

	ListItems(ctx context.Context, accountID string, sinceLesson int) ([]Item, error)
	UpsertItems(ctx context.Context, accountID string, lessonNumber int, results []ItemResult, now time.Time) error
	MarkReviewed(ctx context.Context, accountID string, texts []string, lessonNumber int, now time.Time) error

	ListPositions(ctx context.Context, accountID string) ([]Position, error)
	SavePosition(ctx context.Context, pos Position) error

	DeleteAccountData(ctx context.Context, accountID string) error
}
