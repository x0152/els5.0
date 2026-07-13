package practice

import (
	"fmt"

	"github.com/els/backend/internal/domain/shared"
)

// Kind is the book slug the practice belongs to; books are dynamic, so any non-empty slug is valid.
type Kind string

func (k Kind) Valid() bool {
	return k != ""
}

// MainVariant is the implicit, undeletable set that ships with the chapter theory.
const MainVariant = "main"

const (
	StatusGenerating = "generating"
	StatusReady      = "ready"
	StatusError      = "error"
)

type Variant struct {
	ID        string
	Kind      Kind
	Number    int
	Title     string
	Exercises string
	Status    string
	Error     string
}

func (v Variant) Validate() error {
	var errs []error
	if !v.Kind.Valid() {
		errs = append(errs, fmt.Errorf("variant.kind: invalid"))
	}
	if v.Number <= 0 {
		errs = append(errs, fmt.Errorf("variant.number: must be > 0"))
	}
	return shared.Validation(errs...)
}

type AnswerState struct {
	Answer      string `json:"answer"`
	Correct     bool   `json:"correct"`
	Correction  string `json:"correction,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type Progress struct {
	Answers   map[string]AnswerState
	Completed bool
}

// Source is the chapter content a variant is generated or checked against.
type Source struct {
	Title     string
	Theory    string
	Exercises string
}

type CheckResult struct {
	Correct     bool
	Correction  string
	Explanation string
}
