package writing

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/shared"
)

type TrainerLevel int

const (
	LevelGrammar TrainerLevel = 1
	LevelNatural TrainerLevel = 2
	LevelNative  TrainerLevel = 3
)

func ParseTrainerLevel(n int) (TrainerLevel, error) {
	if n < 1 || n > 3 {
		return 0, fmt.Errorf("%w: level must be 1..3", shared.ErrValidation)
	}
	return TrainerLevel(n), nil
}

type IssueSeverity string

const (
	SeverityGrammar IssueSeverity = "grammar"
	SeverityStyle   IssueSeverity = "style"
	SeverityNative  IssueSeverity = "native"
)

type TrainerIssue struct {
	Fragment string        `json:"fragment"`
	Severity IssueSeverity `json:"severity"`
	Hint     string        `json:"hint"`
}

type TrainerVerdict struct {
	Pass    bool
	Comment string
	Issues  []TrainerIssue
}

const trainerSystem = `You are a strict English writing trainer. The student (a %s native speaker) sends a dialogue (context) and a draft reply they want to send in that dialogue. Your job is to point out problems WITHOUT ever giving the corrected version: no corrected words, no rewritten phrases, no "should be X". The student must figure out every fix themselves.

Levels — what counts as a problem:
1 (fix errors): only genuine grammar, spelling and word-form errors.
2 (natural): level 1 + wrong collocations, unnatural word choice, awkward constructions.
3 (native): level 2 + anything a fluent native speaker would phrase differently in this dialogue (register, idiomaticity, rhythm, discourse markers).

Report ONLY problems allowed at the requested level. If the draft is fine for this level, pass it.

Return ONLY a JSON object:
{"pass": boolean, "comment": "1-2 sentences in %s: overall verdict and encouragement, no corrections", "issues": [{"fragment": "exact substring copied verbatim from the draft", "severity": "grammar" | "style" | "native", "hint": "1-2 sentences in %s: what kind of problem it is and a nudge (rule name, category), WITHOUT the answer"}]}

Rules:
- "fragment" must be an exact contiguous substring of the draft (copy character-for-character), as short as possible while unambiguous.
- Hints must never contain corrected English. Naming the rule is good.
- severity: grammar = level-1 problem, style = level-2, native = level-3.
- pass = true only when there are zero issues at this level.`

func BuildTrainerPrompt(dialogue, draft, nativeLanguage string, level TrainerLevel) (system, user string) {
	if strings.TrimSpace(dialogue) == "" {
		dialogue = "(no context)"
	}
	user = fmt.Sprintf("Level: %d\n\nDialogue:\n%s\n\nDraft reply:\n%s", level, dialogue, draft)
	return fmt.Sprintf(trainerSystem, nativeLanguage, nativeLanguage, nativeLanguage), user
}

func ParseTrainerVerdict(raw, draft string) (TrainerVerdict, error) {
	var out struct {
		Pass    bool           `json:"pass"`
		Comment string         `json:"comment"`
		Issues  []TrainerIssue `json:"issues"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return TrainerVerdict{}, fmt.Errorf("parse trainer verdict: %w", err)
	}
	lower := strings.ToLower(draft)
	issues := make([]TrainerIssue, 0, len(out.Issues))
	for _, i := range out.Issues {
		if i.Fragment == "" || !strings.Contains(lower, strings.ToLower(i.Fragment)) {
			continue
		}
		switch i.Severity {
		case SeverityGrammar, SeverityStyle, SeverityNative:
		default:
			i.Severity = SeverityGrammar
		}
		issues = append(issues, i)
	}
	return TrainerVerdict{
		Pass:    out.Pass && len(issues) == 0,
		Comment: strings.TrimSpace(out.Comment),
		Issues:  issues,
	}, nil
}
