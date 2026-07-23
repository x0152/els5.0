package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type WarmupItemOutput struct {
	Mode    string `json:"mode" enum:"speak,dictation"`
	Text    string `json:"text"`
	FilmID  string `json:"film_id,omitempty"`
	StartMs int    `json:"start_ms,omitempty"`
	EndMs   int    `json:"end_ms,omitempty"`
}

type WatchOutput struct {
	FilmID  string `json:"film_id"`
	Title   string `json:"title"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
	Recap   string `json:"recap,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type QuestionOutput struct {
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
}

type PhraseOutput struct {
	Text    string `json:"text"`
	FilmID  string `json:"film_id,omitempty"`
	StartMs int    `json:"start_ms,omitempty"`
	EndMs   int    `json:"end_ms,omitempty"`
}

type ReadingOutput struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Words []string `json:"words,omitempty"`
}

type WritingOutput struct {
	Scenario string `json:"scenario"`
	Dialogue string `json:"dialogue"`
}

type GrammarOutput struct {
	Topic     string `json:"topic"`
	Theory    string `json:"theory,omitempty"`
	Exercises string `json:"exercises"`
}

type VocabWordOutput struct {
	Text        string `json:"text"`
	Translation string `json:"translation,omitempty"`
	Definition  string `json:"definition,omitempty"`
	Example     string `json:"example,omitempty"`
}

type StepOutput struct {
	ID        string             `json:"id"`
	Kind      string             `json:"kind" enum:"warmup,watch,questions,speak,dictation,reading,writing,grammar,vocab"`
	Title     string             `json:"title"`
	Done      bool               `json:"done"`
	Score     int                `json:"score"`
	Warmup    []WarmupItemOutput `json:"warmup,omitempty"`
	Watch     *WatchOutput       `json:"watch,omitempty"`
	Questions []QuestionOutput   `json:"questions,omitempty"`
	Phrases   []PhraseOutput     `json:"phrases,omitempty"`
	Reading   *ReadingOutput     `json:"reading,omitempty"`
	Writing   *WritingOutput     `json:"writing,omitempty"`
	Grammar   *GrammarOutput     `json:"grammar,omitempty"`
	Vocab     []VocabWordOutput  `json:"vocab,omitempty"`
}

type LessonOutput struct {
	ID         string       `json:"id"`
	Number     int          `json:"number"`
	CycleIndex int          `json:"cycle_index"`
	Review     bool         `json:"review"`
	FilmID     string       `json:"film_id,omitempty"`
	StartMs    int          `json:"start_ms"`
	EndMs      int          `json:"end_ms"`
	Status     string       `json:"status" enum:"active,completed"`
	Steps      []StepOutput `json:"steps"`
	CreatedAt  string       `json:"created_at"`
}

type TodayInput struct {
	authx.BearerInput
}

type WorkoutTodayOutput struct {
	Streak           int           `json:"streak"`
	Days             []string      `json:"days"`
	Completed        bool          `json:"completed"`
	Lesson           *LessonOutput `json:"lesson,omitempty"`
	Generating       bool          `json:"generating,omitempty"`
	GeneratingSince  string        `json:"generating_since,omitempty"`
	GenerationFailed bool          `json:"generation_failed,omitempty"`
}

type StartLessonOutput struct {
	Generating bool          `json:"generating"`
	Lesson     *LessonOutput `json:"lesson,omitempty"`
}

type StartLessonInput struct {
	authx.BearerInput
}

type GetLessonInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type ItemResultInput struct {
	Kind    string `json:"kind" enum:"phrase,word"`
	Text    string `json:"text" minLength:"1" maxLength:"300"`
	FilmID  string `json:"film_id,omitempty"`
	StartMs int    `json:"start_ms,omitempty"`
	EndMs   int    `json:"end_ms,omitempty"`
	Score   int    `json:"score" minimum:"0" maximum:"100"`
}

type SubmitStepInput struct {
	authx.BearerInput
	ID     string `path:"id" format:"uuid"`
	StepID string `path:"step"`
	Body   struct {
		Score   int               `json:"score" minimum:"0" maximum:"100"`
		Results []ItemResultInput `json:"results,omitempty" maxItems:"30"`
	}
}

type ResetInput struct {
	authx.BearerInput
}

type ResetOutput struct{}
