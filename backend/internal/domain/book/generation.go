package book

import (
	"encoding/json"
	"fmt"
	"strings"
)

func BuildChapterPrompt(bookSlug, topic string) (system, user string) {
	system = grammarChapterSystem
	switch bookSlug {
	case "essentialbook":
		system = wordsChapterSystem
	case "wordbook":
		system = vocabularyChapterSystem
	case "phrasebook":
		system = collocationsChapterSystem
	}
	user = fmt.Sprintf("TOPIC: %s", topic)
	return system, user
}

func ParseGeneratedChapter(bookSlug, raw string) (Chapter, error) {
	var out struct {
		Title     string   `json:"title"`
		Words     []string `json:"words"`
		Theory    string   `json:"theory"`
		Exercises string   `json:"exercises"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Chapter{}, fmt.Errorf("parse chapter: %w", err)
	}
	c := Chapter{
		Book:      bookSlug,
		Title:     strings.TrimSpace(out.Title),
		Words:     out.Words,
		Theory:    strings.TrimSpace(out.Theory),
		Exercises: strings.TrimSpace(out.Exercises),
	}
	if c.Words == nil {
		c.Words = []string{}
	}
	if c.Title == "" || c.Theory == "" || c.Exercises == "" {
		return Chapter{}, fmt.Errorf("empty chapter")
	}
	return c, nil
}

const dslReference = `Content DSL (Markdown). THEORY and EXERCISES are markdown strings.

Theory blocks you can use:
- Headings "## A", "## B — Forms", short explanatory paragraphs.
- "**bold**" and "_italic_" inline.
- Examples as "- " bullet lists and "> " blockquotes.
- Tables with the standard markdown "| col | col |" syntax.
- Image placeholder (NOT generated, just a prompt): "~~~image right\n<one English scene description>\n~~~" (side can be right or left).
- Speech bubble: "~~~bubble\n@Name\ntext\n~~~".
- Note box: "~~~box\ntext\n~~~".
- Word bank chip block: "~~~bank\nword, word, word\n~~~".
- Timeline: "~~~timeline\ntick: 0 past\ntick: 100 now\narrow: 30 100 | label\n~~~".
- Glossed reading text (a short passage where key terms carry numbered definitions): "~~~gloss\nRyan **put forward**[1] a **business plan**[2].\n---\n1. offered for people to consider\n2. a plan of how the business will operate\n~~~" — put [N] right after the bold term.
- Collocation fork (stem word + words it combines with): "~~~fork\na remarkable\nrange, likeness, coincidence\n~~~".

Exercise blocks. Each exercise starts with "## <n> → <L>" (n = running number, L = the theory section letter) and an instruction line, then:
- Fill-in gap "{{answer|alt1|alt2}}" — list every acceptable answer separated by "|".
- Multiple choice "{{*correct|wrong1|wrong2}}" — exactly one option prefixed with "*".
- Numbered items "1. ... {{...}} ..." with optional hint "_(hint)_".
- Word bank "~~~bank\nword, word\n~~~".
- Picture prompts "~~~images\none English scene per line\n~~~" used with numbered gap items.
- Matching "~~~match\n1. left :: a\n---\na. right\n~~~".
- Sorting into categories "~~~sort\ncategory one: item, item\ncategory two: item, item\n~~~" — each line is a category with its correct items; the learner sorts a shuffled pool.
- Find-in-text "~~~highlight\nA passage where the learner taps ==each target phrase== hidden ==in running text==.\n~~~" — wrap every phrase to find in "==...==".
- Complete-the-fork "~~~fork\na remarkable\n{{*range|height|speed}}, coincidence\n~~~" — branches may contain gaps; every fork gap MUST be multiple-choice or backed by a "~~~bank" with the answers (a bare gap is unanswerable). One fork per block.
- Dialogue with gaps: several consecutive "~~~bubble\n@Name\nreply with {{gap|alt}}\n~~~" blocks (one per reply, two speakers) form a chat the learner completes.
- Free writing "~~~write rows=N\n> example answer\n~~~".`

const grammarChapterSystem = `You write one complete unit for an English grammar textbook (English Grammar in Use style) on the TOPIC the user gives.
All generated text MUST be in English only.
Produce both the THEORY (clear explanation with examples, organised into sections ## A, ## B, …) and a set of EXERCISES that practise it. You decide how many sections and exercises, and which exercise types to mix — every exercise must clearly practise the theory and be solvable with correct English.

` + dslReference + `

Rules:
- Keep English correct and unambiguous; every gap must have well-defined answers.
- Use image placeholders where a picture helps (they stay as prompts; do not worry about generation).
- Respond with STRICT JSON only: {"title": "<short unit title>", "theory": "<theory markdown>", "exercises": "<exercises markdown>"}.`

const vocabularyChapterSystem = `You write one complete unit for an upper-intermediate English vocabulary textbook (English Vocabulary in Use style) on the TOPIC the user gives.
All generated text MUST be in English only.
Teach vocabulary IN CONTEXT: organise THEORY into sections ## A, ## B, … where words appear inside short readable texts, tables or diagrams — not as a bare word list. Prefer a "~~~gloss~~~" passage (a short text where key terms carry numbered definitions) for at least one section; use markdown tables to group word families or register (formal/informal), and "~~~fork~~~" for words that combine with a common stem. Then produce EXERCISES that practise the words: mix gap-fills, sorting into categories, find-in-text, matching and one free-writing task.

` + dslReference + `

Rules:
- Keep English correct and unambiguous; every gap must have well-defined answers.
- Use image placeholders where a picture helps (they stay as prompts; do not worry about generation).
- Respond with STRICT JSON only: {"title": "<short unit title>", "words": ["word", "word", …], "theory": "<theory markdown>", "exercises": "<exercises markdown>"}.`

const collocationsChapterSystem = `You write one complete unit for an intermediate English collocations textbook (English Collocations in Use style) on the TOPIC the user gives.
All generated text MUST be in English only.
Teach word combinations, not single words: organise THEORY into sections ## A, ## B, … around collocation patterns (verb+noun, adjective+noun, adverb+verb…). Show each collocation inside an example sentence with the combination in **bold**; use "~~~fork~~~" diagrams for a stem with its collocates, markdown tables for verb/collocates-with/example rows, and a "~~~box~~~" for common learner errors (say make a mistake, NOT do a mistake). Then produce EXERCISES: at least one "~~~highlight~~~" find-the-collocations-in-text task, plus gap-fills, sorting, complete-the-fork and matching.

` + dslReference + `

Rules:
- Keep English correct and unambiguous; every gap must have well-defined answers.
- Use image placeholders where a picture helps (they stay as prompts; do not worry about generation).
- Respond with STRICT JSON only: {"title": "<short unit title>", "words": ["collocation", "collocation", …], "theory": "<theory markdown>", "exercises": "<exercises markdown>"}.`

const wordsChapterSystem = `You write one complete vocabulary lesson for an English "Essential Words" textbook on the TOPIC the user gives.
All generated text MUST be in English only.
Pick a coherent set of useful words for the topic. Produce THEORY that teaches each word (meaning, an example sentence, organised into sections ## A, ## B, …) and EXERCISES that practise the words. Mix exercise types; every exercise must clearly practise the words and be solvable with correct English.

` + dslReference + `

Rules:
- Keep English correct and unambiguous; every gap must have well-defined answers.
- Use image placeholders where a picture helps (they stay as prompts; do not worry about generation).
- Respond with STRICT JSON only: {"title": "<short lesson title>", "words": ["word", "word", …], "theory": "<theory markdown>", "exercises": "<exercises markdown>"}.`
