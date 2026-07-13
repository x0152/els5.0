package api

import (
	"time"

	authx "github.com/els/backend/internal/utils/auth"
)

type EventEnvelope struct {
	ClientID   string         `json:"client_id,omitempty"`
	Skill      string         `json:"skill,omitempty" doc:"reading|writing|speaking|listening (optional for pure self-assessment)"`
	Text       string         `json:"text,omitempty" doc:"language sample to decompose (free input)"`
	Target     string         `json:"target,omitempty" doc:"word or grammar concept practiced, plain language (targeted input)"`
	Outcome    string         `json:"outcome,omitempty" doc:"ok|fail (for targeted input)"`
	Context    string         `json:"context,omitempty"`
	Source     map[string]any `json:"source,omitempty" doc:"provenance: app, book_id, video_id, ..."`
	Meta       map[string]any `json:"meta,omitempty" doc:"free bag, e.g. precise internal keys"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
}

type IngestInput struct {
	authx.BearerInput
	Body struct {
		Events []EventEnvelope `json:"events" minItems:"1"`
	}
}

type IngestOutput struct {
	Accepted int `json:"accepted"`
}

type MarkUnclearInput struct {
	authx.BearerInput
	Body EventEnvelope
}

type MarkUnclearOutput struct {
	Updated bool `json:"updated"`
}

type ListInput struct {
	authx.BearerInput
	Status string `query:"status" enum:"pending,processed,failed,all,raw" default:"pending"`
}

type EventView struct {
	ID         string         `json:"id"`
	RawEventID string         `json:"raw_event_id,omitempty"`
	ClientID   string         `json:"client_id,omitempty"`
	Status     string         `json:"status"`
	Skill      string         `json:"skill,omitempty"`
	Text       string         `json:"text,omitempty"`
	Target     string         `json:"target,omitempty"`
	Context    string         `json:"context,omitempty"`
	Action     string         `json:"action,omitempty"`
	Lemma      string         `json:"lemma,omitempty"`
	POS        string         `json:"pos,omitempty"`
	GrammarKey string         `json:"grammar_key,omitempty"`
	Outcome    string         `json:"outcome,omitempty"`
	Error      map[string]any `json:"error,omitempty"`
	Source     map[string]any `json:"source,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
	OccurredAt time.Time      `json:"occurred_at"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ListOutput struct {
	Events []EventView `json:"events"`
}

type CatalogInput struct {
	authx.BearerInput
}

type WordView struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Lemma     string    `json:"lemma"`
	POS       string    `json:"pos,omitempty"`
	Type      string    `json:"type,omitempty"`
	CEFR      string    `json:"cefr,omitempty"`
	Frequency float64   `json:"frequency,omitempty"`
	Enriched  bool      `json:"enriched"`
	CreatedAt time.Time `json:"created_at"`
}

type GrammarRuleView struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	ParentKey string    `json:"parent_key,omitempty"`
	Title     string    `json:"title,omitempty"`
	CEFRLevel string    `json:"cefr_level,omitempty"`
	Enriched  bool      `json:"enriched"`
	CreatedAt time.Time `json:"created_at"`
}

type CatalogOutput struct {
	Words []WordView        `json:"words"`
	Rules []GrammarRuleView `json:"rules"`
}

type DictionariesInput struct {
	authx.BearerInput
}

type DictEntryView struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

type DictionariesOutput struct {
	Dictionaries map[string][]DictEntryView `json:"dictionaries"`
}

type WipeInput struct {
	authx.BearerInput
}

type WipeOutput struct {
	OK bool `json:"ok"`
}

type DeleteRowsInput struct {
	authx.BearerInput
	Body struct {
		Kind string   `json:"kind" enum:"events,raw,words,rules"`
		IDs  []string `json:"ids" minItems:"1"`
	}
}

type DeleteRowsOutput struct {
	Deleted int64 `json:"deleted"`
}
