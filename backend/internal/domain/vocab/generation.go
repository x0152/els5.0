package vocab

import (
	"encoding/json"
	"fmt"
	"strings"
)

type CheckResult struct {
	Correct       bool
	Correction    string
	Explanation   string
	Kind          string
	Text          string
	Transcription string
	Translation   string
	Definition    string
	Example       string
	Frequency     int
	Cefr          string
}

func BuildCheckPrompt(input, nativeLanguage string) (system, user string) {
	return fmt.Sprintf(checkSystem, nativeLanguage, nativeLanguage, nativeLanguage), fmt.Sprintf("ITEM:\n%s", input)
}

func ParseCheckResult(raw string) (CheckResult, error) {
	var out struct {
		Correct       bool   `json:"correct"`
		Correction    string `json:"correction"`
		Explanation   string `json:"explanation"`
		Kind          string `json:"kind"`
		Text          string `json:"text"`
		Transcription string `json:"transcription"`
		Translation   string `json:"translation"`
		Definition    string `json:"definition"`
		Example       string `json:"example"`
		Frequency     int    `json:"frequency"`
		Cefr          string `json:"cefr"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return CheckResult{}, fmt.Errorf("parse vocab check: %w", err)
	}
	return CheckResult(out), nil
}

type PracticeCheckResult struct {
	Correct     bool
	Correction  string
	Explanation string
}

const practiceProductionWords = 3

func wordsList(units []Unit) string {
	var b strings.Builder
	b.WriteString("WORDS:\n")
	for i, u := range units {
		line := fmt.Sprintf("%d. %s", i+1, u.Text)
		if u.Kind != "" {
			line += fmt.Sprintf(" (%s)", u.Kind)
		}
		if u.Translation != "" {
			line += fmt.Sprintf(" — translation: %s", u.Translation)
		}
		if u.Definition != "" {
			line += fmt.Sprintf("; EN: %s", u.Definition)
		}
		if u.Example != "" {
			line += fmt.Sprintf("; example: %s", u.Example)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func BuildMatchPrompt(units []Unit) (system, user string) {
	return matchSystem, wordsList(units)
}

func BuildGapPrompt(units []Unit) (system, user string) {
	return gapSystem, wordsList(units)
}

// BuildWriteExercises produces the writing stage locally (no LLM), numbered from start.
func BuildWriteExercises(units []Unit, start int) string {
	prod := units
	if len(prod) > practiceProductionWords {
		prod = prod[:practiceProductionWords]
	}
	var b strings.Builder
	for i, u := range prod {
		fmt.Fprintf(&b, "## %d\nWrite your own sentence using \"%s\".\n~~~write rows=2\n~~~\n\n", start+i, u.Text)
	}
	return strings.TrimSpace(b.String())
}

func ParseGeneratedPractice(raw string) (string, error) {
	var out struct {
		Exercises string `json:"exercises"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", fmt.Errorf("parse practice: %w", err)
	}
	exercises := strings.TrimSpace(out.Exercises)
	if exercises == "" {
		return "", fmt.Errorf("empty practice")
	}
	return exercises, nil
}

func BuildPracticeCheckPrompt(instruction, answer string) (system, user string) {
	return practiceCheckSystem, fmt.Sprintf("TASK:\n%s\n\nSTUDENT ANSWER:\n%s", instruction, answer)
}

func ParsePracticeCheckResult(raw string) (PracticeCheckResult, error) {
	var out struct {
		Correct     bool   `json:"correct"`
		Correction  string `json:"correction"`
		Explanation string `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return PracticeCheckResult{}, fmt.Errorf("parse practice check: %w", err)
	}
	return PracticeCheckResult(out), nil
}

type AnalyzeItem struct {
	Text        string   `json:"text"`
	Kind        string   `json:"kind"`
	Lemmas      []string `json:"lemmas"`
	Description string   `json:"description"`
	Translation string   `json:"translation"`
	Frequency   int      `json:"frequency"`
	Cefr        string   `json:"cefr"`
}

func BuildAnalyzePrompt(selection, context, level, nativeLanguage string) (system, user string) {
	system = fmt.Sprintf(analyzeSystem, nativeLanguage)
	if strings.TrimSpace(level) != "" {
		system += fmt.Sprintf("\n\nThe learner's English level is %s — match the wording of the description to that level.", strings.TrimSpace(level))
	}
	if context == "" {
		return system, fmt.Sprintf("SELECTION:\n%s", selection)
	}
	return system, fmt.Sprintf("CONTEXT:\n%s\n\nSELECTION:\n%s", context, selection)
}

func ParseAnalyzeResult(raw string) ([]AnalyzeItem, error) {
	raw = strings.TrimSpace(raw)
	var wrapped struct {
		Items []AnalyzeItem `json:"items"`
	}
	if json.Unmarshal([]byte(raw), &wrapped) == nil && len(wrapped.Items) > 0 {
		return wrapped.Items, nil
	}
	items := make([]AnalyzeItem, 0)
	for _, line := range strings.Split(raw, "\n") {
		if it, ok := ParseAnalyzeLine(line); ok {
			items = append(items, it)
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("parse vocab analyze: no items")
	}
	return items, nil
}

func ParseAnalyzeLine(line string) (AnalyzeItem, bool) {
	line = strings.TrimSpace(line)
	if line == "" || line[0] != '{' {
		return AnalyzeItem{}, false
	}
	var it AnalyzeItem
	if json.Unmarshal([]byte(line), &it) != nil || strings.TrimSpace(it.Text) == "" {
		return AnalyzeItem{}, false
	}
	return it, true
}

const analyzeSystem = `You help an English student turn a fragment of English text they selected while reading or watching into vocabulary items to memorize.
You receive the selected text and optionally the surrounding context (use context only to disambiguate meaning, never analyze words outside the selection).

Break the SELECTION into EVERY distinct memorizable item it contains, without limiting their number: every content word (noun, verb, adjective, adverb) and every set phrase, phrasal verb and idiom. Be exhaustive — cover the whole selection, not just a few highlights. Prefer the most useful learning unit: if several words form a phrasal verb or idiom, return that as ONE item instead of separate words; otherwise return each meaningful word as its own item. Skip only standalone function words (articles, pronouns, prepositions, conjunctions, auxiliary/modal verbs). Do not skip a word just because it looks common.
When the selection is more than one word, ALWAYS also include one item for the WHOLE selection as a single phrase, explained in its general dictionary meaning independent of this particular context.

For each item fill:
- "text": the normalized English item (lowercase unless a proper noun), in dictionary form.
- "kind": one of "word", "phrase", "phrasal_verb", "idiom".
- "lemmas": the dictionary base forms to look the item up by. For a single word give its lemma; for a phrasal verb give the verb and particle as one entry like "look up"; for a phrase or idiom give the lemmatized whole phrase. Lowercase, no punctuation.
- "description": a short, simple English explanation of the meaning in this context (one short phrase, plain words). The description itself must stay in English.
- "translation": a concise %s translation of the item as used in this context.
- "frequency": an integer 1–5 for how common the item is in everyday English overall (5 = extremely common core word everyone knows, 3 = average, 1 = rare, literary or specialized).
- "cefr": the CEFR level at which a learner usually meets this item — exactly one of "A1", "A2", "B1", "B2", "C1", "C2".

Respond as JSON Lines: output ONE item per line as a compact JSON object, nothing else — no surrounding array, no wrapping object, no code fences, no commentary.
Each line must be exactly: {"text": "", "kind": "", "lemmas": [""], "description": "", "translation": "", "frequency": 0, "cefr": ""}`

const checkSystem = `You help a %s-speaking student build a personal English vocabulary collection for memorization.
The student submits ONE item to add: a word, a phrase, a phrasal verb, or an idiom. The input may contain typos or be written partly in %s as a request.

First decide whether it is a valid, correctly spelled English vocabulary item.
- If it is misspelled, invalid, or not English, set "correct" to false, put the corrected English form in "correction" and a short English explanation of what is wrong in "explanation". Leave the remaining fields empty.
- If it is correct, set "correct" to true and fill:
  - "kind": one of "word", "phrase", "phrasal_verb", "idiom".
  - "text": the normalized item (lowercase unless it is a proper noun).
  - "transcription": IPA transcription for a single word, empty for multi-word items.
  - "translation": a concise %s translation.
  - "definition": a clear English definition.
  - "example": one natural English example sentence using the item.
  - "frequency": an integer 1–5 for how common the item is in everyday English overall (5 = extremely common core word everyone knows, 3 = average, 1 = rare, literary or specialized).
  - "cefr": the CEFR level at which a learner usually meets this item — exactly one of "A1", "A2", "B1", "B2", "C1", "C2".

Respond with STRICT JSON only: {"correct": <true|false>, "correction": "", "explanation": "", "kind": "", "text": "", "transcription": "", "translation": "", "definition": "", "example": "", "frequency": 0, "cefr": ""}.`

const matchSystem = `You create ONE matching exercise for a vocabulary practice worksheet, based ONLY on the given list of words.
All generated text MUST be in English.
Output ONLY this Markdown DSL block, starting with a "## 1" header:

## 1
<a short instruction line>
~~~match
1. <word> :: <letter>
(one line per word, in the given order)
---
a. <definition>
(one definition per word, lettered a, b, c, …, in SHUFFLED order)
~~~
Each "::" letter must point to the line of that word's CORRECT definition. Cover ALL the words exactly once.
Respond with STRICT JSON only: {"exercises": "<the markdown block>"}.`

const gapSystem = `You create ONE gap-fill exercise for a vocabulary practice worksheet, based ONLY on the given list of words.
All generated text MUST be in English.
Output ONLY this Markdown DSL block, starting with a "## 2" header:

## 2
<a short instruction line>
~~~bank
<word>, <word>, … (all words, comma-separated)
~~~
1. <a fresh, natural sentence with the target word replaced by {{<word>}}>
(one numbered sentence per word, covering ALL words exactly once)
Write NEW sentences. Each sentence has exactly one gap and a single correct answer.
Respond with STRICT JSON only: {"exercises": "<the markdown block>"}.`

const practiceCheckSystem = `You are an English teacher checking a student's sentence written for a vocabulary task.
Judge whether the sentence fulfils the task (uses the target word correctly) AND is grammatically correct English.
Be encouraging but accurate; minor typos are fine if meaning and grammar are clear.
Respond with STRICT JSON only: {"correct": <true|false>, "correction": "<a corrected version of the sentence, empty if already correct>", "explanation": "<a short English explanation of what is right or wrong>"}.`
