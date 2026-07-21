package workout

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/els/backend/internal/domain/shared"
)

const (
	LessonStatusActive    = "active"
	LessonStatusCompleted = "completed"
)

const CycleLength = 7

const (
	StepWarmup    = "warmup"
	StepWatch     = "watch"
	StepQuestions = "questions"
	StepSpeak     = "speak"
	StepDictation = "dictation"
	StepReading   = "reading"
	StepWriting   = "writing"
	StepGrammar   = "grammar"
	StepVocab     = "vocab"
)

var stepSkills = map[string]string{
	StepWarmup:    "review",
	StepWatch:     "listening",
	StepQuestions: "listening",
	StepSpeak:     "speaking",
	StepDictation: "listening",
	StepReading:   "reading",
	StepWriting:   "writing",
	StepGrammar:   "grammar",
	StepVocab:     "vocab",
}

func StepSkill(kind string) string { return stepSkills[kind] }

type Step struct {
	ID      string          `json:"id"`
	Kind    string          `json:"kind"`
	Title   string          `json:"title"`
	Payload json.RawMessage `json:"payload"`
	Done    bool            `json:"done"`
	Score   int             `json:"score"`
}

type Lesson struct {
	ID          string
	AccountID   string
	Number      int
	FilmID      string
	StartMs     int
	EndMs       int
	Status      string
	Steps       []Step
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// CycleIndex is the lesson's position inside its 7-lesson cycle; index 7 is the review lesson.
func (l Lesson) CycleIndex() int { return (l.Number-1)%CycleLength + 1 }

func IsReviewNumber(number int) bool { return number%CycleLength == 0 }

func (l *Lesson) SubmitStep(stepID string, score int, now time.Time) error {
	if l.Status == LessonStatusCompleted {
		return fmt.Errorf("lesson is already completed: %w", shared.ErrConflict)
	}
	found := false
	allDone := true
	for i := range l.Steps {
		if l.Steps[i].ID == stepID {
			l.Steps[i].Done = true
			l.Steps[i].Score = clampScore(score)
			found = true
		}
		if !l.Steps[i].Done {
			allDone = false
		}
	}
	if !found {
		return fmt.Errorf("step %q: %w", stepID, shared.ErrNotFound)
	}
	if allDone {
		l.Status = LessonStatusCompleted
		l.CompletedAt = &now
	}
	return nil
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// Step payloads: each step is self-contained, the lesson never regenerates content at play time.

type WarmupItem struct {
	Mode    string `json:"mode" enum:"speak,dictation"`
	Text    string `json:"text"`
	FilmID  string `json:"film_id,omitempty"`
	StartMs int    `json:"start_ms,omitempty"`
	EndMs   int    `json:"end_ms,omitempty"`
}

type WarmupPayload struct {
	Items []WarmupItem `json:"items"`
}

type WatchPayload struct {
	FilmID  string `json:"film_id"`
	Title   string `json:"title"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
	Recap   string `json:"recap"`
	Summary string `json:"summary"`
}

type Question struct {
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
}

type QuestionsPayload struct {
	Questions []Question `json:"questions"`
}

type SpeakPhrase struct {
	Text    string `json:"text"`
	FilmID  string `json:"film_id,omitempty"`
	StartMs int    `json:"start_ms,omitempty"`
	EndMs   int    `json:"end_ms,omitempty"`
}

type SpeakPayload struct {
	Phrases []SpeakPhrase `json:"phrases"`
}

type DictationPayload struct {
	Sentences []SpeakPhrase `json:"sentences"`
}

type ReadingPayload struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Words []string `json:"words"`
}

type WritingPayload struct {
	Scenario string `json:"scenario"`
	Dialogue string `json:"dialogue"`
}

type GrammarPayload struct {
	Topic     string `json:"topic"`
	Exercises string `json:"exercises"`
}

type VocabWord struct {
	Text        string `json:"text"`
	Translation string `json:"translation,omitempty"`
	Definition  string `json:"definition,omitempty"`
	Example     string `json:"example,omitempty"`
}

type VocabPayload struct {
	Words []VocabWord `json:"words"`
}

func NewStep(id, kind, title string, payload any) (Step, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Step{}, fmt.Errorf("marshal %s payload: %w", kind, err)
	}
	return Step{ID: id, Kind: kind, Title: title, Payload: raw}, nil
}

// Streak counts consecutive calendar days with a completed lesson, ending today or yesterday.
// Dates must be unique days sorted descending.
func Streak(dates []time.Time, today time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	cursor := today
	if !sameDay(dates[0], cursor) {
		cursor = cursor.AddDate(0, 0, -1)
		if !sameDay(dates[0], cursor) {
			return 0
		}
	}
	streak := 0
	for _, d := range dates {
		if !sameDay(d, cursor) {
			break
		}
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

func sameDay(a, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}
