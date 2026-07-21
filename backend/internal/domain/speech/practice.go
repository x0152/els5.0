package speech

import (
	"encoding/json"
	"fmt"
	"strings"
)

const practiceSystem = `You create short pronunciation practice sentences for an English learner (B1-B2). Produce 4 natural sentences (8-14 words each) that are comfortable to read aloud: everyday vocabulary, no tongue-twisters, no rare proper names.

If target sounds (IPA) are given, pack each sentence with common words containing those sounds.

Return ONLY a JSON object:
{"sentences": ["...", "...", "...", "..."]}`

func BuildPracticePrompt(topic string, sounds []string) (system, user string) {
	var b strings.Builder
	if strings.TrimSpace(topic) == "" {
		b.WriteString("Topic: pick everyday situations yourself.")
	} else {
		fmt.Fprintf(&b, "Topic: %s", topic)
	}
	if len(sounds) > 0 {
		fmt.Fprintf(&b, "\n\nTarget sounds to practice: /%s/", strings.Join(sounds, "/, /"))
	}
	return practiceSystem, b.String()
}

func ParsePractice(raw string) ([]string, error) {
	var out struct {
		Sentences []string `json:"sentences"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse practice: %w", err)
	}
	sentences := make([]string, 0, len(out.Sentences))
	for _, s := range out.Sentences {
		if s = strings.TrimSpace(s); s != "" {
			sentences = append(sentences, s)
		}
	}
	if len(sentences) == 0 {
		return nil, fmt.Errorf("parse practice: no sentences")
	}
	return sentences, nil
}
