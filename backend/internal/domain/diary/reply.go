package diary

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Reply struct {
	Text         string
	NextQuestion string
	NativeSample string
	Corrections  []Correction
}

const replySystem = `You are a warm, attentive friend and gentle psychologist inside an English-learning diary app. The user (a %s native speaker learning English) writes a short diary entry in English, answering today's question.

You do TWO independent jobs in one response:

1. FRIEND REPLY (field "reply", in English, simple and warm):
- React to the CONTENT like a close friend: empathize, reflect, notice what matters. 2-4 sentences.
- Never mention language mistakes here. Never correct anything here.
- End with "next_question": one short, personal follow-up question about their life that naturally continues the conversation and gives them a topic for tomorrow's entry.

2. LANGUAGE EDITOR (silent, separate fields):
- "corrections": the 2-5 most valuable language issues. For each: "sentence" — the user's sentence exactly as written; "fragment" — the exact wrong substring copied verbatim from that sentence; "correction" — the corrected fragment; "description" — one short explanation in %s (the user's native language).
- Skip trivial typos unless they change meaning. Prioritize repeated or fossilized errors.
- "native_sample": rewrite the user's whole entry (2-4 sentences) the way a fluent native speaker would naturally say the same thing. Keep their meaning and personality.

Return ONLY a JSON object:
{"reply": "...", "next_question": "...", "corrections": [{"sentence": "...", "fragment": "...", "correction": "...", "description": "..."}], "native_sample": "..."}`

func BuildReplyPrompt(question, text, nativeLanguage string, history []Entry) (system, user string) {
	var b strings.Builder
	if len(history) > 0 {
		b.WriteString("Recent entries for context (oldest first):\n")
		for _, h := range history {
			fmt.Fprintf(&b, "- %s: %s\n", h.Date.Format("2006-01-02"), truncate(h.Text, 300))
		}
		b.WriteString("\n")
	}
	if question != "" {
		fmt.Fprintf(&b, "Today's question:\n%s\n\n", question)
	}
	fmt.Fprintf(&b, "Today's entry:\n%s", text)
	return fmt.Sprintf(replySystem, nativeLanguage, nativeLanguage), b.String()
}

func ParseReply(raw string) (Reply, error) {
	var out struct {
		Reply        string       `json:"reply"`
		NextQuestion string       `json:"next_question"`
		Corrections  []Correction `json:"corrections"`
		NativeSample string       `json:"native_sample"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Reply{}, fmt.Errorf("parse diary reply: %w", err)
	}
	if strings.TrimSpace(out.Reply) == "" {
		return Reply{}, fmt.Errorf("parse diary reply: empty reply")
	}
	corrections := make([]Correction, 0, len(out.Corrections))
	for _, c := range out.Corrections {
		if strings.TrimSpace(c.Fragment) == "" {
			continue
		}
		corrections = append(corrections, c)
	}
	return Reply{
		Text:         strings.TrimSpace(out.Reply),
		NextQuestion: strings.TrimSpace(out.NextQuestion),
		NativeSample: strings.TrimSpace(out.NativeSample),
		Corrections:  corrections,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
