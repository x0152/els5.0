package studio_test

import (
	"errors"
	"testing"
	"time"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/studio"
)

func TestReviewSchedule(t *testing.T) {
	// arrange
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	item := studio.Item{Listened: true, Spoken: true, Written: true, Recalled: true}

	// act: mastering schedules the first review in 2 days
	item.ScheduleReviewIfDone(now)

	// assert
	if item.ReviewStage != 1 || item.NextReviewAt == nil || !item.NextReviewAt.Equal(now.AddDate(0, 0, 2)) {
		t.Fatalf("first review: stage=%d next=%v", item.ReviewStage, item.NextReviewAt)
	}
	if item.ReviewDue(now) {
		t.Fatal("review must not be due before next_review_at")
	}

	// act: pass reviews at +2, +7, +30 days
	for i, days := range []int{7, 30} {
		now = *item.NextReviewAt
		if !item.ReviewDue(now) {
			t.Fatalf("review %d must be due", i+2)
		}
		if err := item.PassReview(now); err != nil {
			t.Fatalf("pass review %d: %v", i+2, err)
		}
		if item.NextReviewAt == nil || !item.NextReviewAt.Equal(now.AddDate(0, 0, days)) {
			t.Fatalf("review %d: next=%v want +%dd", i+2, item.NextReviewAt, days)
		}
	}

	// act: passing the last review finishes the cycle
	now = *item.NextReviewAt
	if err := item.PassReview(now); err != nil {
		t.Fatalf("final review: %v", err)
	}
	if item.NextReviewAt != nil {
		t.Fatalf("after final review next must be nil, got %v", item.NextReviewAt)
	}

	// assert: passing without a due review is a validation error
	if err := item.PassReview(now); !errors.Is(err, shared.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestScheduleReviewIfDoneOnlyOnce(t *testing.T) {
	// arrange
	now := time.Now()
	item := studio.Item{Listened: true, Spoken: true, Written: true, Recalled: true}
	item.ScheduleReviewIfDone(now)
	first := *item.NextReviewAt

	// act
	item.ScheduleReviewIfDone(now.AddDate(0, 0, 1))

	// assert
	if item.ReviewStage != 1 || !item.NextReviewAt.Equal(first) {
		t.Fatalf("schedule must not advance twice: stage=%d next=%v", item.ReviewStage, item.NextReviewAt)
	}
}
