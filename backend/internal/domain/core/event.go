package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusProcessed  Status = "processed"
	StatusFailed     Status = "failed"
	StatusAll        Status = "all"
	StatusRaw        Status = "raw"
)

func ParseStatus(s string) (Status, error) {
	switch Status(s) {
	case StatusPending, StatusProcessed, StatusFailed, StatusAll, StatusRaw:
		return Status(s), nil
	default:
		return "", fmt.Errorf("%w: status: unknown value %q", shared.ErrValidation, s)
	}
}

const (
	SkillReading   = "reading"
	SkillWriting   = "writing"
	SkillSpeaking  = "speaking"
	SkillListening = "listening"
)

var skills = map[string]bool{
	SkillReading:   true,
	SkillWriting:   true,
	SkillSpeaking:  true,
	SkillListening: true,
}

func ValidSkill(s string) bool { return skills[s] }

// IsProductive reports whether the skill is one the learner produces (writing,
// speaking) — those carry both vocabulary and errors; receptive skills don't.
func IsProductive(skill string) bool { return skill == SkillWriting || skill == SkillSpeaking }

type RawEvent struct {
	ID          string
	UserID      string
	ClientID    string
	Status      string
	Skill       string
	Text        string
	Target      string
	Outcome     string
	Context     string
	Source      json.RawMessage
	Meta        json.RawMessage
	OccurredAt  time.Time
	CreatedAt   time.Time
	ProcessedAt *time.Time
	Error       string
}

// IsTargeted reports whether the event references a concrete item the learner
// practiced (a word or grammar concept) — those become events directly, while
// free text is decomposed by the LLM.
func (e RawEvent) IsTargeted() bool { return strings.TrimSpace(e.Target) != "" }

func (e RawEvent) Validate() error {
	if e.IsTargeted() {
		if e.Skill != "" && !ValidSkill(e.Skill) {
			return fmt.Errorf("%w: unknown skill %q", shared.ErrValidation, e.Skill)
		}
		return nil
	}
	if !ValidSkill(e.Skill) {
		return fmt.Errorf("%w: skill is required for free input", shared.ErrValidation)
	}
	if strings.TrimSpace(e.Text) == "" {
		return fmt.Errorf("%w: text is required for free input", shared.ErrValidation)
	}
	return nil
}

type EventError struct {
	Name        string `json:"name,omitempty"`
	Sentence    string `json:"sentence,omitempty"`
	Fragment    string `json:"fragment,omitempty"`
	Correction  string `json:"correction,omitempty"`
	Description string `json:"description,omitempty"`
}

type Event struct {
	ID         string
	RawEventID string
	UserID     string
	ClientID   string
	Type       string
	Action     string
	Unit       string
	Lemma      string
	POS        string
	GrammarKey string
	Outcome    string
	Error      *EventError
	Context    string
	Source     json.RawMessage
	Meta       json.RawMessage
	OccurredAt time.Time
	CreatedAt  time.Time
}

type Extraction struct {
	Action     string      `json:"action"`
	Unit       string      `json:"unit"`
	Lemma      string      `json:"lemma"`
	POS        string      `json:"pos"`
	GrammarKey string      `json:"grammar_key"`
	Outcome    string      `json:"outcome"`
	Error      *EventError `json:"error"`
	Example    *EventError `json:"example"`
}

func Normalize(e *RawEvent, now time.Time) {
	e.Skill = strings.ToLower(strings.TrimSpace(e.Skill))
	e.Outcome = strings.ToLower(strings.TrimSpace(e.Outcome))
	e.Target = strings.TrimSpace(e.Target)
	e.Status = string(StatusPending)
	if e.OccurredAt.IsZero() {
		e.OccurredAt = now
	}
	e.CreatedAt = now
	if len(e.Source) == 0 {
		e.Source = json.RawMessage("{}")
	}
	if len(e.Meta) == 0 {
		e.Meta = json.RawMessage("{}")
	}
}
