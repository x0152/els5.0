package vocab

import (
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared"
)

type Kind string

const (
	KindWord        Kind = "word"
	KindPhrase      Kind = "phrase"
	KindPhrasalVerb Kind = "phrasal_verb"
	KindIdiom       Kind = "idiom"
)

func (k Kind) IsValid() bool {
	switch k {
	case KindWord, KindPhrase, KindPhrasalVerb, KindIdiom:
		return true
	}
	return false
}

type Status string

const (
	StatusNew      Status = "new"
	StatusLearning Status = "learning"
	StatusLearned  Status = "learned"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusNew, StatusLearning, StatusLearned:
		return true
	}
	return false
}

type Unit struct {
	ID             string
	AccountID      string
	Text           string
	Kind           Kind
	Transcription  string
	Translation    string
	Definition     string
	Example        string
	Frequency      int
	CEFR           string
	Status         Status
	CorrectStreak  int
	LastAnsweredAt *time.Time
	CreatedAt      time.Time
}

func NewUnit(id, accountID string, c CheckResult) (Unit, error) {
	u := Unit{
		ID:            id,
		AccountID:     accountID,
		Text:          strings.TrimSpace(c.Text),
		Kind:          Kind(strings.TrimSpace(c.Kind)),
		Transcription: strings.TrimSpace(c.Transcription),
		Translation:   strings.TrimSpace(c.Translation),
		Definition:    strings.TrimSpace(c.Definition),
		Example:       strings.TrimSpace(c.Example),
		Frequency:     ClampFrequency(c.Frequency),
		CEFR:          NormalizeCEFR(c.Cefr),
		Status:        StatusNew,
	}
	if err := u.validate(); err != nil {
		return Unit{}, err
	}
	return u, nil
}

var cefrLevels = map[string]bool{"A1": true, "A2": true, "B1": true, "B2": true, "C1": true, "C2": true}

func ClampFrequency(f int) int {
	if f < 0 {
		return 0
	}
	if f > 5 {
		return 5
	}
	return f
}

func NormalizeCEFR(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if cefrLevels[s] {
		return s
	}
	return ""
}

func (u Unit) validate() error {
	var errs []error
	if u.ID == "" {
		errs = append(errs, fmt.Errorf("unit.id: must not be empty"))
	}
	if u.AccountID == "" {
		errs = append(errs, fmt.Errorf("unit.account_id: must not be empty"))
	}
	if u.Text == "" {
		errs = append(errs, fmt.Errorf("unit.text: must not be empty"))
	}
	if !u.Kind.IsValid() {
		errs = append(errs, fmt.Errorf("unit.kind: invalid %q", u.Kind))
	}
	return shared.Validation(errs...)
}
