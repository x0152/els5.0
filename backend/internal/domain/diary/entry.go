package diary

import (
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared"
)

type Correction struct {
	Sentence    string `json:"sentence"`
	Fragment    string `json:"fragment"`
	Correction  string `json:"correction"`
	Description string `json:"description"`
}

type Entry struct {
	ID           string
	AccountID    string
	Date         time.Time
	Question     string
	Draft        string
	Text         string
	Reply        string
	NextQuestion string
	NativeSample string
	Corrections  []Correction
	CreatedAt    time.Time
}

func (e Entry) Validate() error {
	if strings.TrimSpace(e.Text) == "" {
		return fmt.Errorf("%w: text is required", shared.ErrValidation)
	}
	if e.AccountID == "" {
		return fmt.Errorf("%w: account is required", shared.ErrValidation)
	}
	return nil
}

// SameDay compares calendar days ignoring time and timezone of the instants.
func SameDay(a, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

// Streak counts consecutive diary days ending today or yesterday.
// Dates must be unique calendar days sorted descending.
func Streak(dates []time.Time, today time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	cursor := today
	if !SameDay(dates[0], cursor) {
		cursor = cursor.AddDate(0, 0, -1)
		if !SameDay(dates[0], cursor) {
			return 0
		}
	}
	streak := 0
	for _, d := range dates {
		if !SameDay(d, cursor) {
			break
		}
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

var defaultQuestions = []string{
	"What made you smile today, and why did it matter to you?",
	"What is one thing you keep postponing, and what would change if you finally did it?",
	"Describe a small moment from today that you would like to remember in a year.",
	"What was the most difficult part of your day, and how did you handle it?",
	"If tomorrow had one extra free hour, how would you spend it?",
	"What is something you learned recently that surprised you?",
	"Who did you talk to today, and what was the conversation about?",
}

func DefaultQuestion(today time.Time) string {
	return defaultQuestions[today.YearDay()%len(defaultQuestions)]
}
