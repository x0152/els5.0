package writing

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Situation struct {
	Scenario string
	Dialogue string
}

const situationSystem = `You invent realistic everyday situations for an English learner (B1-B2) to practice writing replies: chats with friends, work messages, emails, customer support, small talk.

Create a short dialogue (2-4 turns, speakers labeled like "Anna:" / "You wrote earlier:" etc.) that ends with a message addressed to the learner, so a reply is naturally expected. Keep it lively and specific.

Return ONLY a JSON object:
{"scenario": "one short sentence in English describing the setting", "dialogue": "the dialogue, one turn per line"}`

func BuildSituationPrompt(topic string) (system, user string) {
	if strings.TrimSpace(topic) == "" {
		return situationSystem, "Topic: pick an unexpected everyday situation yourself."
	}
	return situationSystem, fmt.Sprintf("Topic: %s", topic)
}

func ParseSituation(raw string) (Situation, error) {
	var out struct {
		Scenario string `json:"scenario"`
		Dialogue string `json:"dialogue"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Situation{}, fmt.Errorf("parse situation: %w", err)
	}
	if strings.TrimSpace(out.Dialogue) == "" {
		return Situation{}, fmt.Errorf("parse situation: empty dialogue")
	}
	return Situation{Scenario: strings.TrimSpace(out.Scenario), Dialogue: strings.TrimSpace(out.Dialogue)}, nil
}
