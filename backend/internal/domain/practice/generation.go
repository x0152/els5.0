package practice

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PlanItem is one outlined exercise: which theory section it practises and its type.
type PlanItem struct {
	Section string `json:"section"`
	Type    string `json:"type"`
	Hint    string `json:"hint"`
}

const maxPlanItems = 8

func BuildPlanPrompt(src Source) (system, user string) {
	user = fmt.Sprintf("THEORY:\n%s", src.Theory)
	return planSystem, user
}

func ParsePlan(raw string) (title string, items []PlanItem, err error) {
	var out struct {
		Title string     `json:"title"`
		Items []PlanItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", nil, fmt.Errorf("parse plan: %w", err)
	}
	if len(out.Items) == 0 {
		return "", nil, fmt.Errorf("empty plan")
	}
	if len(out.Items) > maxPlanItems {
		out.Items = out.Items[:maxPlanItems]
	}
	return strings.TrimSpace(out.Title), out.Items, nil
}

func BuildExercisePrompt(src Source, item PlanItem, number int) (system, user string) {
	user = fmt.Sprintf("THEORY:\n%s\n\nONE EXISTING EXERCISE SET (only as a DSL syntax reference — do NOT copy its items):\n%s\n\nCREATE EXERCISE #%d:\n- theory section: %s\n- type: %s\n- focus: %s",
		src.Theory, src.Exercises, number, item.Section, item.Type, item.Hint)
	return exerciseSystem, user
}

func ParseGeneratedExercise(raw string) (string, error) {
	var out struct {
		Exercise string `json:"exercise"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", fmt.Errorf("parse exercise: %w", err)
	}
	ex := strings.TrimSpace(out.Exercise)
	if ex == "" {
		return "", fmt.Errorf("empty exercise")
	}
	return ex, nil
}

func BuildCheckPrompt(theory, instruction, answer string) (system, user string) {
	user = fmt.Sprintf("THEORY (for reference):\n%s\n\nTASK:\n%s\n\nSTUDENT ANSWER:\n%s", theory, instruction, answer)
	return checkSystem, user
}

func ParseCheckResult(raw string) (CheckResult, error) {
	var out struct {
		Correct     bool   `json:"correct"`
		Correction  string `json:"correction"`
		Explanation string `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return CheckResult{}, fmt.Errorf("parse check: %w", err)
	}
	return CheckResult{Correct: out.Correct, Correction: out.Correction, Explanation: out.Explanation}, nil
}

const planSystem = `You plan a fresh set of practice exercises for an English textbook (English Grammar in Use / Essential English Words style) from the chapter THEORY.
Decide a varied outline: how many exercises (4 to 7) and which TYPES to use, in a sensible order — mix them, do not reuse the same shape every time. Each exercise must practise the given theory.
For each planned exercise give:
- "section": the theory section letter it relates to (e.g. "A").
- "type": one of "gap-fill", "multiple-choice", "word-bank", "matching", "sorting", "find-in-text", "dialogue", "picture", "free-writing".
- "hint": a short note (in English) on what exactly this exercise should drill.
Respond with STRICT JSON only: {"title": "<short label, e.g. 'Extra practice'>", "items": [{"section": "", "type": "", "hint": ""}]}.`

const exerciseSystem = `You write ONE practice exercise for an English textbook from the chapter THEORY, following the requested type and focus.
All generated text MUST be in English only. The existing exercise set is shown ONLY as a DSL syntax reference.

Output the exercise as a Markdown block that starts with a header line "## <n> → <L>" (the given number and theory section letter), followed by an instruction line, then the body using these building blocks (pick what fits the type):
- Fill-in gap: "{{answer|alt1|alt2}}" — list every acceptable answer separated by "|".
- Multiple choice: "{{*correct|wrong1|wrong2}}" — exactly one option, the correct one, is prefixed with "*".
- Numbered items: lines like "1. ... {{...}} ...". A small hint can follow as "_(hint)_".
- Word bank: a "~~~bank\nword, word, word\n~~~" block of helper words.
- Picture prompts: a "~~~images\none English scene description per line\n~~~" block (used with numbered gap items that describe each picture).
- Matching: "~~~match\n1. left text :: a\n...\n---\na. right text\n...\n~~~".
- Sorting into categories: "~~~sort\ncategory one: item, item\ncategory two: item, item\n~~~" — each line is a category followed by its correct items; the learner sorts a shuffled pool.
- Find-in-text: "~~~highlight\nA short passage where the learner taps ==each target phrase== hidden ==in running text==.\n~~~" — wrap every phrase the learner must find in "==...==".
- Dialogue with gaps: several consecutive "~~~bubble\n@Name\nreply text with {{gap|alt}}\n~~~" blocks (one block per reply, two speakers) form a chat where the learner completes the replies.
- Collocation fork with gaps: "~~~fork\na remarkable\n{{*range|height|speed}}, coincidence\n~~~" — first line is the stem, branches may contain gaps. A bare text gap in a branch is unanswerable (too many words fit): every fork gap MUST be multiple-choice ({{*correct|wrong|wrong}}) or the exercise MUST include a "~~~bank" with the answers. One fork per "~~~fork" block; use several blocks for several stems.
- Free writing: "~~~write rows=N\n> an example answer\n~~~".
- Inline "**bold**" and "_italic_" are allowed.

Rules:
- Output ONLY this one exercise (no theory, no other exercises, no explanations).
- Keep English correct and unambiguous; gaps must have well-defined answers.
- Respond with STRICT JSON only: {"exercise": "<the exercise markdown>"}.`

const checkSystem = `You are an English teacher checking a student's free-form written answer to a task.
Judge whether the answer fulfils the task AND is grammatically correct English.
Be encouraging but accurate; minor typos are acceptable if meaning and grammar are clear.
Respond with STRICT JSON only: {"correct": <true|false>, "correction": "<a corrected version of the answer, empty if already correct>", "explanation": "<a short explanation in English of what is right or wrong>"}.`
