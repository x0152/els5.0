package listening

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/shared"
)

type Dictation struct {
	Sentences []string
}

type Level string

const (
	LevelEasy   Level = "easy"
	LevelMedium Level = "medium"
	LevelHard   Level = "hard"
)

func ParseLevel(s string) (Level, error) {
	switch Level(s) {
	case "":
		return LevelMedium, nil
	case LevelEasy, LevelMedium, LevelHard:
		return Level(s), nil
	}
	return "", fmt.Errorf("%w: level must be easy, medium or hard", shared.ErrValidation)
}

const (
	MinSentences = 3
	MaxSentences = 10
)

func ParseSentenceCount(n int) (int, error) {
	if n == 0 {
		return 5, nil
	}
	if n < MinSentences || n > MaxSentences {
		return 0, fmt.Errorf("%w: count must be %d..%d", shared.ErrValidation, MinSentences, MaxSentences)
	}
	return n, nil
}

var levelSpec = map[Level]string{
	LevelEasy:   "an A2-B1 learner: short simple sentences (6-10 words), high-frequency vocabulary, no idioms",
	LevelMedium: "a B1-B2 learner: natural spoken sentences (8-16 words) with contractions and common collocations",
	LevelHard:   "a B2-C1 learner: longer natural sentences (12-20 words) with phrasal verbs, idioms and connected speech",
}

const dictationSystem = `You create dictation exercises for %s. Produce %d standalone sentences of natural spoken English on the given topic: everyday speech, no rare proper names or numbers that are hard to spell.

If a list of learner's words is provided, use some of them where they fit naturally.

Return ONLY a JSON object:
{"sentences": ["...", "...", "..."]}`

func BuildDictationPrompt(topic string, words []string, level Level, count int) (system, user string) {
	var b strings.Builder
	if strings.TrimSpace(topic) == "" {
		b.WriteString("Topic: pick everyday situations yourself, each sentence may differ.")
	} else {
		fmt.Fprintf(&b, "Topic: %s", topic)
	}
	if len(words) > 0 {
		fmt.Fprintf(&b, "\n\nLearner's words to use when natural: %s", strings.Join(words, ", "))
	}
	return fmt.Sprintf(dictationSystem, levelSpec[level], count), b.String()
}

func ParseDictation(raw string) (Dictation, error) {
	var out struct {
		Sentences []string `json:"sentences"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Dictation{}, fmt.Errorf("parse dictation: %w", err)
	}
	sentences := make([]string, 0, len(out.Sentences))
	for _, s := range out.Sentences {
		if s = strings.TrimSpace(s); s != "" {
			sentences = append(sentences, s)
		}
	}
	if len(sentences) == 0 {
		return Dictation{}, fmt.Errorf("parse dictation: no sentences")
	}
	return Dictation{Sentences: sentences}, nil
}
