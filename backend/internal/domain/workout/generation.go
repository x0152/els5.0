package workout

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/films"
)

const questionsSystem = `You write comprehension questions about a scene the learner has just watched, based on its subtitles.
Write %d multiple-choice questions in simple English (%s learner): what happened, who said what, why. Questions must be answerable from the scene alone.
Each question has exactly 4 options with a single correct one; distractors must be plausible.

Return ONLY a JSON object:
{"questions": [{"text": "...", "options": ["...", "...", "...", "..."], "answer": 0}]}`

func BuildQuestionsPrompt(recap string, cues []films.Cue, userLevel string, count int) (system, user string) {
	var b strings.Builder
	if strings.TrimSpace(recap) != "" {
		fmt.Fprintf(&b, "Previously: %s\n\n", recap)
	}
	b.WriteString("Scene subtitles:\n")
	for _, c := range cues {
		b.WriteString(strings.ReplaceAll(c.Text, "\n", " "))
		b.WriteString("\n")
	}
	return fmt.Sprintf(questionsSystem, count, NormalizeLevel(userLevel)), b.String()
}

func ParseQuestions(raw string) ([]Question, error) {
	var out struct {
		Questions []Question `json:"questions"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse questions: %w", err)
	}
	questions := make([]Question, 0, len(out.Questions))
	for _, q := range out.Questions {
		if strings.TrimSpace(q.Text) == "" || len(q.Options) != 4 || q.Answer < 0 || q.Answer > 3 {
			continue
		}
		questions = append(questions, q)
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("parse questions: no valid questions")
	}
	return questions, nil
}

// GrammarFocus is one recurring learner error the grammar drill targets.
type GrammarFocus struct {
	Rule     string
	Sentence string
	Hint     string
}

const grammarSystem = `You write a full grammar unit in the style of Murphy's "English Grammar in Use" for a %s English learner. All text in English only.

THEORY (the left page of a Murphy unit, generous and friendly):
- 2-4 lettered sections ("## A — <aspect>", "## B — <aspect>", ...), each explaining one aspect of the rule in short, simple sentences a learner actually understands: when it is used, how it is formed, what it contrasts with.
- Every section gives 2-4 example sentences as a bulleted list with the target form in **bold**.
- Include a "~~~box\n<the key rule in one or two memorable lines>\n~~~" callout.
- Include one or two "~~~image right md\n<a vivid everyday scene illustrating an example sentence, described in one line>\n~~~" illustrations next to the examples they depict.
- Where the rule contrasts two forms, show a small comparison as two bulleted groups.

EXERCISES (the right page of a Murphy unit):
- 5-6 exercises of DIFFERENT shapes ("## 1 → A" headers: the running number and the letter of the theory section it practises; one instruction line each, 4-6 items per exercise), from easy recognition to harder production, all practising this rule, built from these blocks:
%s

Rules:
- Do not use the free-writing block.
- Keep English correct and unambiguous; gaps must have well-defined answers.

Return ONLY a JSON object: {"topic": "<short label of the grammar focus>", "theory": "<theory markdown>", "exercises": "<exercises markdown>"}`

// BuildGrammarPrompt targets the learner's recent mistakes; without any it falls back to
// grammar that is typically hard at the learner's level, themed around the lesson topic.
// blockCatalog is the exercise DSL reference (practice.BlockCatalog).
func BuildGrammarPrompt(focuses []GrammarFocus, topic, userLevel, blockCatalog string) (system, user string) {
	var b strings.Builder
	if len(focuses) == 0 {
		b.WriteString("The learner has no recorded mistakes yet. Pick ONE grammar point that is typically hard at this level.\n")
	} else {
		b.WriteString("Build the unit around ONE grammar point — the one behind most of the learner's recent mistakes below — in fresh sentences (do not copy their sentences):\n")
	}
	for _, f := range focuses {
		fmt.Fprintf(&b, "- rule: %s", f.Rule)
		if f.Sentence != "" {
			fmt.Fprintf(&b, "; their sentence: %q", f.Sentence)
		}
		if f.Hint != "" {
			fmt.Fprintf(&b, "; note: %s", f.Hint)
		}
		b.WriteString("\n")
	}
	if strings.TrimSpace(topic) != "" {
		fmt.Fprintf(&b, "\nIf it helps, theme the sentences around: %s\n", topic)
	}
	return fmt.Sprintf(grammarSystem, NormalizeLevel(userLevel), blockCatalog), b.String()
}

func ParseGrammar(raw string) (GrammarPayload, error) {
	var out GrammarPayload
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return GrammarPayload{}, fmt.Errorf("parse grammar: %w", err)
	}
	if strings.TrimSpace(out.Exercises) == "" {
		return GrammarPayload{}, fmt.Errorf("parse grammar: empty exercises")
	}
	return out, nil
}

// DictationLines picks the longest cues of the watched block as dictation material, keeping
// sentences the learner can actually type (a few words, no music/sound tags).
func DictationLines(cues []films.Cue, filmID string, limit int) []SpeakPhrase {
	candidates := []films.Cue{}
	for _, c := range cues {
		text := strings.TrimSpace(strings.ReplaceAll(c.Text, "\n", " "))
		words := len(strings.Fields(text))
		if words < 5 || words > 16 || strings.ContainsAny(text, "[]♪#") {
			continue
		}
		c.Text = text
		candidates = append(candidates, c)
	}
	step := 1
	if len(candidates) > limit && limit > 0 {
		step = len(candidates) / limit
	}
	out := []SpeakPhrase{}
	for i := 0; i < len(candidates) && len(out) < limit; i += step {
		c := candidates[i]
		out = append(out, SpeakPhrase{Text: c.Text, FilmID: filmID, StartMs: c.StartMs, EndMs: c.EndMs})
	}
	return out
}
