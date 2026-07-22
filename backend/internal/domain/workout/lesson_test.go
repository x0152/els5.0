package workout_test

import (
	"errors"
	"testing"
	"time"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/workout"
)

func day(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestStreak(t *testing.T) {
	cases := []struct {
		name  string
		dates []string
		today string
		want  int
	}{
		{"empty", nil, "2026-07-18", 0},
		{"today only", []string{"2026-07-18"}, "2026-07-18", 1},
		{"ends yesterday", []string{"2026-07-17", "2026-07-16"}, "2026-07-18", 2},
		{"broken two days ago", []string{"2026-07-16"}, "2026-07-18", 0},
		{"gap inside", []string{"2026-07-18", "2026-07-16"}, "2026-07-18", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			dates := make([]time.Time, 0, len(tc.dates))
			for _, d := range tc.dates {
				dates = append(dates, day(d))
			}
			// act
			got := workout.Streak(dates, day(tc.today))
			// assert
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSubmitStep(t *testing.T) {
	now := day("2026-07-18")

	t.Run("unknown step", func(t *testing.T) {
		// arrange
		lesson := workout.Lesson{Status: workout.LessonStatusActive, Steps: []workout.Step{{ID: "s1"}}}
		// act
		err := lesson.SubmitStep("nope", 50, now)
		// assert
		if !errors.Is(err, shared.ErrNotFound) {
			t.Fatalf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("completed lesson rejects", func(t *testing.T) {
		// arrange
		lesson := workout.Lesson{Status: workout.LessonStatusCompleted, Steps: []workout.Step{{ID: "s1"}}}
		// act
		err := lesson.SubmitStep("s1", 50, now)
		// assert
		if !errors.Is(err, shared.ErrConflict) {
			t.Fatalf("got %v, want ErrConflict", err)
		}
	})

	t.Run("last step completes the lesson", func(t *testing.T) {
		// arrange
		lesson := workout.Lesson{Status: workout.LessonStatusActive, Steps: []workout.Step{{ID: "s1", Done: true}, {ID: "s2"}}}
		// act
		if err := lesson.SubmitStep("s2", 120, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// assert
		if lesson.Status != workout.LessonStatusCompleted || lesson.CompletedAt == nil {
			t.Fatalf("lesson not completed: %+v", lesson)
		}
		if lesson.Steps[1].Score != 100 {
			t.Fatalf("score not clamped: %d", lesson.Steps[1].Score)
		}
	})

	t.Run("middle step keeps lesson active", func(t *testing.T) {
		// arrange
		lesson := workout.Lesson{Status: workout.LessonStatusActive, Steps: []workout.Step{{ID: "s1"}, {ID: "s2"}}}
		// act
		if err := lesson.SubmitStep("s1", 80, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// assert
		if lesson.Status != workout.LessonStatusActive {
			t.Fatalf("lesson unexpectedly completed")
		}
	})
}

func TestGenerationInFlight(t *testing.T) {
	now := time.Now()
	fresh := workout.Lesson{Status: workout.LessonStatusGenerating, CreatedAt: now.Add(-time.Minute)}
	stale := workout.Lesson{Status: workout.LessonStatusGenerating, CreatedAt: now.Add(-workout.GenerationStaleAfter - time.Minute)}
	failed := workout.Lesson{Status: workout.LessonStatusFailed, CreatedAt: now}
	if !fresh.GenerationInFlight(now) || stale.GenerationInFlight(now) || failed.GenerationInFlight(now) {
		t.Fatal("generation in-flight wrong")
	}
}

func TestCycle(t *testing.T) {
	if (workout.Lesson{Number: 7}).CycleIndex() != 7 || (workout.Lesson{Number: 8}).CycleIndex() != 1 {
		t.Fatal("cycle index wrong")
	}
	if !workout.IsReviewNumber(7) || !workout.IsReviewNumber(14) || workout.IsReviewNumber(6) {
		t.Fatal("review number wrong")
	}
}
