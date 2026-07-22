package studio

import (
	"encoding/json"
	"fmt"
	"strings"
)

const enrichSystem = `You are a language assistant in an English-learning app. The user (a %s native speaker) added a word or phrase to their study list.

Return ONLY a JSON object:
{
  "transcription": "IPA transcription of the text, without slashes",
  "translation": "natural translation into %s",
  "explanation": "one short simple-English sentence explaining what the text means and when to use it",
  "explanation_native": "the same explanation translated into %s",
  "example": "one short natural everyday sentence using the text exactly as given"
}`

type Enrichment struct {
	Transcription     string `json:"transcription"`
	Translation       string `json:"translation"`
	Explanation       string `json:"explanation"`
	ExplanationNative string `json:"explanation_native"`
	Example           string `json:"example"`
}

func BuildEnrichPrompt(text, nativeLanguage string) (system, user string) {
	return fmt.Sprintf(enrichSystem, nativeLanguage, nativeLanguage, nativeLanguage), "TEXT:\n" + text
}

func ParseEnrichment(raw string) (Enrichment, error) {
	var out Enrichment
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Enrichment{}, fmt.Errorf("parse studio enrichment: %w", err)
	}
	return out, nil
}

const exampleSystem = `You are a language assistant in an English-learning app. Write ONE short natural everyday sentence that uses the given word or phrase exactly as given. It must differ from the previous example.

Return ONLY a JSON object: {"example": "..."}`

func BuildExamplePrompt(text, previous string) (system, user string) {
	return exampleSystem, fmt.Sprintf("TEXT:\n%s\n\nPREVIOUS EXAMPLE:\n%s", text, previous)
}

func ParseExample(raw string) (string, error) {
	var out struct {
		Example string `json:"example"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", fmt.Errorf("parse studio example: %w", err)
	}
	if strings.TrimSpace(out.Example) == "" {
		return "", fmt.Errorf("parse studio example: empty example")
	}
	return strings.TrimSpace(out.Example), nil
}

const taskSystem = `You are a language coach in an English-learning app. The user is studying a word or phrase. Invent ONE short real-life mini-situation (1-2 sentences, in simple English) where replying naturally requires using that word or phrase. Address the user directly, e.g. "A colleague says: ... — reply using the phrase." It must differ from the previous task.

Return ONLY a JSON object: {"task": "..."}`

func BuildTaskPrompt(text, previous string) (system, user string) {
	return taskSystem, fmt.Sprintf("TEXT:\n%s\n\nPREVIOUS TASK:\n%s", text, previous)
}

func ParseTask(raw string) (string, error) {
	var out struct {
		Task string `json:"task"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", fmt.Errorf("parse studio task: %w", err)
	}
	if strings.TrimSpace(out.Task) == "" {
		return "", fmt.Errorf("parse studio task: empty task")
	}
	return strings.TrimSpace(out.Task), nil
}

const checkSystem = `You are a language coach in an English-learning app. The user is studying the word or phrase given below. They received a mini-situation and wrote a reply. Judge the reply:
- ok=true only if the reply uses the studied word or phrase (small inflections are fine) naturally and without real grammar mistakes;
- otherwise ok=false.
"comment": one or two short encouraging sentences in simple English — praise what works, name what to fix. Never write the full corrected reply for the user.

Return ONLY a JSON object: {"ok": true|false, "comment": "..."}`

type ReplyCheck struct {
	OK      bool   `json:"ok"`
	Comment string `json:"comment"`
}

func BuildCheckPrompt(text, task, reply string) (system, user string) {
	return checkSystem, fmt.Sprintf("STUDIED PHRASE:\n%s\n\nSITUATION:\n%s\n\nUSER REPLY:\n%s", text, task, reply)
}

func ParseCheck(raw string) (ReplyCheck, error) {
	var out ReplyCheck
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return ReplyCheck{}, fmt.Errorf("parse studio check: %w", err)
	}
	return out, nil
}
