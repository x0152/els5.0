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
	Draft        string             `json:"draft,omitempty"`
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
		Draft    string `json:"draft,omitempty" maxLength:"5000" doc:"The first version of the text, before error fixes"`
	}
}

type CheckEntryInput struct {
	authx.BearerInput
	Body struct {
		Text string `json:"text" minLength:"1" maxLength:"5000" doc:"Diary draft to check for grammar errors"`
	}
}

type GrammarErrorOutput struct {
	Original    string `json:"original"`
	Correction  string `json:"correction"`
	Explanation string `json:"explanation"`
	Type        string `json:"type"`
}

type CheckEntryOutput struct {
	OK     bool                 `json:"ok"`
	Errors []GrammarErrorOutput `json:"errors"`
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
