package speech

import "context"

type Verdict string

const (
	VerdictGood    Verdict = "good"
	VerdictClose   Verdict = "close"
	VerdictWrong   Verdict = "wrong"
	VerdictMissing Verdict = "missing"
)

type Phoneme struct {
	Expected string
	Heard    string
	Score    float64
	Verdict  Verdict
}

type Word struct {
	Word     string
	IPA      string
	Score    int
	Phonemes []Phoneme
	Extra    []string
}

type Assessment struct {
	Overall int
	Heard   string
	Words   []Word
}

type Assessor interface {
	Assess(ctx context.Context, audio []byte, text string, strictness float64) (Assessment, error)
}

const (
	MinStrictness     = 0.5
	MaxStrictness     = 2.5
	DefaultStrictness = 1.0
)
