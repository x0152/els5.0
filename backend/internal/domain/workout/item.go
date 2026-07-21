package workout

import (
	"sort"
	"time"
)

const (
	ItemPhrase = "phrase"
	ItemWord   = "word"
)

// Item is one unit of cycle material (a phrase or word the learner met) with its latest result;
// the spiral picks weak items back into warm-ups and review lessons.
type Item struct {
	ID            string
	AccountID     string
	Kind          string
	Text          string
	FilmID        string
	StartMs       int
	EndMs         int
	LessonNumber  int
	LastScore     int
	TimesReviewed int
	LastLesson    int
	UpdatedAt     time.Time
}

type ItemResult struct {
	Kind    string
	Text    string
	FilmID  string
	StartMs int
	EndMs   int
	Score   int
}

// PickWarmup selects spiral material for lesson warm-up: items introduced 1 or 3 lessons
// ago plus anything weak, worst first, capped at limit.
func PickWarmup(items []Item, lessonNumber, limit int) []Item {
	picked := []Item{}
	for _, it := range items {
		age := lessonNumber - it.LessonNumber
		if age == 1 || age == 3 || it.LastScore < 60 {
			picked = append(picked, it)
		}
	}
	sortWeakFirst(picked)
	if len(picked) > limit {
		picked = picked[:limit]
	}
	return picked
}

// PickReview selects the weakest material of the finishing cycle for the review lesson.
func PickReview(items []Item, lessonNumber, limit int) []Item {
	picked := []Item{}
	for _, it := range items {
		if lessonNumber-it.LessonNumber < CycleLength {
			picked = append(picked, it)
		}
	}
	sortWeakFirst(picked)
	if len(picked) > limit {
		picked = picked[:limit]
	}
	return picked
}

func sortWeakFirst(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].LastScore != items[j].LastScore {
			return items[i].LastScore < items[j].LastScore
		}
		return items[i].TimesReviewed < items[j].TimesReviewed
	})
}
