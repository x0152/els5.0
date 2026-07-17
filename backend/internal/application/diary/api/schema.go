package api

import (
	"time"

	authx "github.com/els/backend/internal/utils/auth"
)

type CorrectionOutput struct {
	Sentence    string `json:"sentence"`
	Fragment    string `json:"fragment"`
	Correction  string `json:"correction"`
	Description string `json:"description"`
}

type EntryOutput struct {
	ID           string             `json:"id"`
	Date         string             `json:"date"`
	Question     string             `json:"question,omitempty"`
	Text         string             `json:"text"`
	Reply        string             `json:"reply"`
	NextQuestion string             `json:"next_question,omitempty"`
	NativeSample string             `json:"native_sample,omitempty"`
	Corrections  []CorrectionOutput `json:"corrections"`
	CreatedAt    time.Time          `json:"created_at"`
}

type GetTodayInput struct {
	authx.BearerInput
}

type TodayOutput struct {
	Question string             `json:"question"`
	Entry    *EntryOutput       `json:"entry,omitempty"`
	Warmup   []CorrectionOutput `json:"warmup"`
	Streak   int                `json:"streak"`
}

type SubmitEntryInput struct {
	authx.BearerInput
	Body struct {
		Text     string `json:"text" minLength:"1" maxLength:"5000" doc:"Diary entry text in English"`
		Question string `json:"question,omitempty" maxLength:"500" doc:"The question the entry answers"`
	}
}

type ListEntriesInput struct {
	authx.BearerInput
	Limit  int32 `query:"limit" minimum:"1" maximum:"100" default:"30"`
	Offset int32 `query:"offset" minimum:"0" default:"0"`
}

type EntriesOutput struct {
	Items  []EntryOutput `json:"items"`
	Total  int64         `json:"total"`
	Limit  int32         `json:"limit"`
	Offset int32         `json:"offset"`
}

type ResetHistoryInput struct {
	authx.BearerInput
}

type ResetHistoryOutput struct{}

type TrainerCheckInput struct {
	authx.BearerInput
	Body struct {
		Dialogue string `json:"dialogue,omitempty" maxLength:"3000" doc:"Dialogue context the draft replies to"`
		Draft    string `json:"draft" minLength:"1" maxLength:"2000" doc:"Draft reply to check"`
		Level    int    `json:"level" minimum:"1" maximum:"3" doc:"Strictness: 1 grammar, 2 natural, 3 native"`
	}
}

type TrainerIssueOutput struct {
	Fragment string `json:"fragment"`
	Severity string `json:"severity" enum:"grammar,style,native"`
	Hint     string `json:"hint"`
}

type TrainerCheckOutput struct {
	Pass    bool                 `json:"pass"`
	Comment string               `json:"comment"`
	Issues  []TrainerIssueOutput `json:"issues"`
}
