package vocab

import "context"

const (
	PracticeStatusGenerating = "generating"
	PracticeStatusReady      = "ready"
	PracticeStatusError      = "error"
)

type PracticeAnswer struct {
	Answer      string `json:"answer"`
	Correct     bool   `json:"correct"`
	Correction  string `json:"correction,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type PracticeSession struct {
	ID        string
	Words     []Unit
	Exercises string
	Status    string
	Error     string
	Answers   map[string]PracticeAnswer
	Completed bool
}

// PracticeSessionRepository stores the learner's single latest practice session.
type PracticeSessionRepository interface {
	Load(ctx context.Context, accountID string) (PracticeSession, error)
	Create(ctx context.Context, accountID string, s PracticeSession) error
	AppendExercises(ctx context.Context, accountID, sessionID, chunk, status, errMsg string) error
	SaveProgress(ctx context.Context, accountID, sessionID string, answers map[string]PracticeAnswer, completed bool) error
}
