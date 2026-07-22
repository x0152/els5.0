package studio

import (
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared"
)

const (
	SkillListened = "listened"
	SkillSpoken   = "spoken"
	SkillWritten  = "written"
	SkillRecalled = "recalled"
)

var reviewIntervalDays = []int{2, 7, 30}

type Area struct {
	ID        string
	AccountID string
	Title     string
	Icon      string
	CreatedAt time.Time
}

type AreaStats struct {
	Area
	Total int
	Done  int
	Due   int
}

type Item struct {
	ID                string
	AreaID            string
	AccountID         string
	Text              string
	Transcription     string
	Translation       string
	Explanation       string
	ExplanationNative string
	Example           string
	Task              string
	Listened          bool
	Spoken            bool
	Written           bool
	Recalled          bool
	ReviewStage       int
	NextReviewAt      *time.Time
	CreatedAt         time.Time
}

func (a Area) Validate() error {
	if strings.TrimSpace(a.Title) == "" {
		return fmt.Errorf("%w: title is required", shared.ErrValidation)
	}
	if a.AccountID == "" {
		return fmt.Errorf("%w: account is required", shared.ErrValidation)
	}
	return nil
}

func (i Item) Validate() error {
	if strings.TrimSpace(i.Text) == "" {
		return fmt.Errorf("%w: text is required", shared.ErrValidation)
	}
	if i.AccountID == "" {
		return fmt.Errorf("%w: account is required", shared.ErrValidation)
	}
	return nil
}

func (i *Item) MarkSkill(skill string) error {
	switch skill {
	case SkillListened:
		i.Listened = true
	case SkillSpoken:
		i.Spoken = true
	case SkillWritten:
		i.Written = true
	case SkillRecalled:
		i.Recalled = true
	default:
		return fmt.Errorf("%w: unknown skill %q", shared.ErrValidation, skill)
	}
	return nil
}

func (i Item) Done() bool {
	return i.Listened && i.Spoken && i.Written && i.Recalled
}

func (i Item) ReviewDue(now time.Time) bool {
	return i.NextReviewAt != nil && !now.Before(*i.NextReviewAt)
}

func (i *Item) ScheduleReviewIfDone(now time.Time) {
	if i.Done() && i.ReviewStage == 0 {
		i.advanceReview(now)
	}
}

func (i *Item) PassReview(now time.Time) error {
	if !i.ReviewDue(now) {
		return fmt.Errorf("%w: no review is due", shared.ErrValidation)
	}
	i.advanceReview(now)
	return nil
}

func (i *Item) advanceReview(now time.Time) {
	i.ReviewStage++
	if i.ReviewStage > len(reviewIntervalDays) {
		i.NextReviewAt = nil
		return
	}
	next := now.AddDate(0, 0, reviewIntervalDays[i.ReviewStage-1])
	i.NextReviewAt = &next
}
