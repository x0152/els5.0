package reading

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/shared"
)

type Text struct {
	Title string
	Body  string
	Words []string
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

type Length string

const (
	LengthShort  Length = "short"
	LengthMedium Length = "medium"
	LengthLong   Length = "long"
)

func ParseLength(s string) (Length, error) {
	switch Length(s) {
	case "":
		return LengthMedium, nil
	case LengthShort, LengthMedium, LengthLong:
		return Length(s), nil
	}
	return "", fmt.Errorf("%w: length must be short, medium or long", shared.ErrValidation)
}

var levelSpec = map[Level]string{
	LevelEasy:   "an A2-B1 learner: simple sentences, high-frequency vocabulary, explain anything unusual in easy words",
	LevelMedium: "a B1-B2 learner: natural, contemporary English with a few less common but useful words",
	LevelHard:   "a B2-C1 learner: rich vocabulary, idioms and phrasal verbs, varied sentence structure",
}

var lengthSpec = map[Length]string{
	LengthShort:  "120-180 words",
	LengthMedium: "220-300 words",
	LengthLong:   "380-500 words",
}

const textSystem = `You are a writer creating reading passages for %s. Write an engaging text (%s) on the given topic: a story, an article or a curious explainer.

Illustrations: only where a picture genuinely helps to understand the scene, insert a separate line exactly in the form:
[image: short English description of the picture]
Use 0-2 such lines for the whole text. Never put them mid-sentence.

If a list of learner's words is provided, naturally weave in as many of them as fit the topic (do not force all of them).

Return ONLY a JSON object:
{"title": "short title", "body": "the text as markdown paragraphs separated by blank lines, [image: ...] lines allowed", "words": ["learner words you actually used"]}`

func BuildTextPrompt(topic string, words []string, level Level, length Length) (system, user string) {
	var b strings.Builder
	if strings.TrimSpace(topic) == "" {
		b.WriteString("Topic: pick something unexpected and interesting yourself.")
	} else {
		fmt.Fprintf(&b, "Topic: %s", topic)
	}
	if len(words) > 0 {
		fmt.Fprintf(&b, "\n\nLearner's words to weave in when natural: %s", strings.Join(words, ", "))
	}
	return fmt.Sprintf(textSystem, levelSpec[level], lengthSpec[length]), b.String()
}

func ParseText(raw string) (Text, error) {
	var out struct {
		Title string   `json:"title"`
		Body  string   `json:"body"`
		Words []string `json:"words"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Text{}, fmt.Errorf("parse reading text: %w", err)
	}
	if strings.TrimSpace(out.Body) == "" {
		return Text{}, fmt.Errorf("parse reading text: empty body")
	}
	return Text{Title: strings.TrimSpace(out.Title), Body: strings.TrimSpace(out.Body), Words: out.Words}, nil
}
