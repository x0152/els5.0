package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared/vo"
)

const leniencyRule = `Do NOT report unimportant issues: informal language and slang, abbreviations and shortenings (e.g. "smth", "lang", "repo", "info", "btw", "etc."), spelling typos/mistyping, capitalization, punctuation, or minor article (a/an/the) choices. Capture only errors that genuinely affect grammar or word meaning.
All names, descriptions and explanations you write MUST be in English only.`

const productionTemplate = `You analyze a BATCH of "%s" events: text the learner PRODUCED themselves.
Input is a JSON array of events, each with an integer "index".
Return ONLY one JSON object: {"results": [{"index": <int>, "items": [ ... ]}]} with exactly one entry per input event, matched by its index.
Be EXHAUSTIVE: for each event capture EVERY lexical unit AND EVERY error. Do not skip, summarize, deduplicate, or limit the number of items.
Each item: {"action": string, "unit": string, "lemma": string, "pos": string, "grammar_key": string, "outcome": "ok"|"fail", "error": object|null}.
Lexical units the learner used (outcome "ok", error null, action "%s"):
- "unit": "word" — one item per word: set lemma (base form) and pos; leave grammar_key empty. Include all words, function words too.
- "unit": "phrase"|"phrasal_verb"|"idiom" — one item per multi-word unit: set lemma to the unit in base form; leave pos empty.
Errors the learner made (outcome "fail", unit empty) — only MEANINGFUL ones:
- "action": "grammar_error" — set grammar_key; leave lemma/pos empty unless one specific word is wrong.
- "action": "word_error" — set lemma and pos of the misused word.
- error: {"name": string, "sentence": string, "fragment": string, "correction": string, "description": string} — name is a short label, sentence is the full original sentence, fragment is the exact wrong part copied verbatim, correction is the fixed fragment, description explains what is wrong.
` + leniencyRule + `
grammar_key format is "category:subtype" optionally ":voice", lowercase snake_case, NO spaces (e.g. "subject_verb_agreement:present_simple").`

const immersionTemplate = `You analyze a BATCH of "%s" events: text the learner encountered or produced.
Input is a JSON array of events, each with an integer "index".
Return ONLY one JSON object: {"results": [{"index": <int>, "items": [ ... ]}]} with exactly one entry per input event, matched by its index.
Be EXHAUSTIVE: for each event capture EVERY word in its text. Do not skip anything, do not summarize, do not deduplicate, do not limit the number of items.
Each item: {"action": "%s", "lemma": string, "pos": string, "outcome": "ok", "error": null}.
- One item for every word in the text: set lemma (base form) and pos. Include all words, function words too.`

const constructionsTemplate = `You extract GRAMMAR CONSTRUCTIONS correctly used in a BATCH of "%s" events.
All names and descriptions you write MUST be in English only.
Input is a JSON array of events, each with an integer "index".
Return ONLY one JSON object: {"results": [{"index": <int>, "items": [ ... ]}]} with exactly one entry per input event, matched by its index.
Be EXHAUSTIVE: capture AS MANY grammar constructions as possible per event — every tense, aspect, voice, mood, article, determiner, clause type, comparison, modality, conditional, question form, etc. Do not skip, summarize, deduplicate, or limit the number of items.
Only real, well-formed grammar constructions. Do NOT report spelling, typos, mistakes, or any error-like keys — this pass is about correct grammar patterns present in the text.
Each item: {"grammar_key": string, "example": {"name": string, "sentence": string, "fragment": string, "description": string}}.
- name: short human label of the construction (e.g. "Present Perfect").
- sentence: the full sentence where it appears, copied from the text.
- fragment: the exact words in that sentence that form the construction, copied verbatim.
- description: one short sentence explaining how the construction is used here.
grammar_key format is "category:subtype" optionally ":voice", lowercase snake_case, NO spaces (e.g. "tense:present_perfect:active", "article:indefinite", "clause:relative:defining").`

func BuildExtractionPrompt(skill string, raws []RawEvent, registry []GrammarRule) (system, user string) {
	return extractionSystem(skill, registry), extractionInputs(raws)
}

func BuildConstructionsPrompt(skill string, raws []RawEvent, registry []GrammarRule) (system, user string) {
	return fmt.Sprintf(constructionsTemplate, skill) + grammarRegistryBlock(registry), extractionInputs(raws)
}

func extractionInputs(raws []RawEvent) string {
	type in struct {
		Index   int    `json:"index"`
		Text    string `json:"text"`
		Context string `json:"context"`
	}
	items := make([]in, len(raws))
	for i, r := range raws {
		items[i] = in{Index: i, Text: r.Text, Context: r.Context}
	}
	b, _ := json.Marshal(items)
	return string(b)
}

func extractionSystem(skill string, registry []GrammarRule) string {
	if IsProductive(skill) {
		return fmt.Sprintf(productionTemplate, skill, ImmersionAction(skill)) + grammarRegistryBlock(registry)
	}
	return fmt.Sprintf(immersionTemplate, skill, ImmersionAction(skill))
}

func WantsGrammarRegistry(skill string) bool { return IsProductive(skill) }

func WantsConstructions(skill string) bool { return true }

func grammarRegistryBlock(registry []GrammarRule) string {
	if len(registry) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nKnown grammar_key registry — reuse an existing key when the mistake matches it; only create a new key when none fits:\n")
	for _, r := range registry {
		b.WriteString("- ")
		b.WriteString(r.Key)
		if r.Title != "" {
			b.WriteString(" — ")
			b.WriteString(r.Title)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func ImmersionAction(skill string) string {
	switch skill {
	case SkillReading:
		return "read"
	case SkillListening:
		return "heard"
	case SkillWriting:
		return "used_in_writing"
	case SkillSpeaking:
		return "used_in_speech"
	default:
		return skill
	}
}

func ParseExtractions(out string) (map[int][]Extraction, error) {
	var parsed struct {
		Results []struct {
			Index int          `json:"index"`
			Items []Extraction `json:"items"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}
	byIndex := make(map[int][]Extraction, len(parsed.Results))
	for _, r := range parsed.Results {
		byIndex[r.Index] = r.Items
	}
	return byIndex, nil
}

// TargetedEvent builds a single structured event for a deliberate practice/review:
// the learner worked on a known word or grammar concept, so no LLM is needed.
func TargetedEvent(raw RawEvent, now time.Time) Event {
	e := newEvent(raw, now)
	e.Outcome = raw.Outcome
	if strings.Contains(raw.Target, ":") {
		e.GrammarKey = NormalizeGrammarKey(raw.Target)
	} else {
		e.Lemma = strings.ToLower(strings.TrimSpace(raw.Target))
	}
	switch raw.Outcome {
	case "ok":
		e.Action = "reviewed_ok"
	case "fail":
		e.Action = "reviewed_fail"
	default:
		e.Action = "reviewed"
	}
	return e
}

func EventsFromExtractions(raw RawEvent, items []Extraction, now time.Time) []Event {
	out := make([]Event, 0, len(items))
	for _, it := range items {
		if it.Action == "" {
			continue
		}
		e := newEvent(raw, now)
		e.Action, e.Unit, e.Lemma, e.POS, e.GrammarKey, e.Outcome, e.Error = it.Action, it.Unit, it.Lemma, it.POS, NormalizeGrammarKey(it.GrammarKey), it.Outcome, it.Error
		out = append(out, e)
	}
	return out
}

func ConstructionEventsFromExtractions(raw RawEvent, items []Extraction, now time.Time) []Event {
	out := make([]Event, 0, len(items))
	for _, it := range items {
		gk := NormalizeGrammarKey(it.GrammarKey)
		if gk == "" {
			continue
		}
		e := newEvent(raw, now)
		e.Action, e.GrammarKey, e.Outcome, e.Error = "construction_used", gk, "ok", it.Example
		out = append(out, e)
	}
	return out
}

func newEvent(raw RawEvent, now time.Time) Event {
	return Event{
		ID:         vo.NewID().String(),
		RawEventID: raw.ID,
		UserID:     raw.UserID,
		ClientID:   raw.ClientID,
		Type:       raw.Skill,
		Context:    raw.Context,
		Source:     raw.Source,
		Meta:       raw.Meta,
		OccurredAt: raw.OccurredAt,
		CreatedAt:  now,
	}
}
