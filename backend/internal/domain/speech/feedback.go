package speech

import (
	"encoding/json"
	"fmt"
	"strings"
)

type FeedbackTip struct {
	Sound  string
	Advice string
}

type Feedback struct {
	Summary string
	Tips    []FeedbackTip
}

const feedbackSystem = `You are an experienced English pronunciation coach.
The student read a text aloud; an acoustic model transcribed what was actually said into IPA and detected per-phoneme issues.
Explain in practical terms what went wrong and how to fix it, considering typical pronunciation habits of native %s speakers.
Focus on the 2-4 most important problems, not every small deviation. Be encouraging but concrete: describe tongue and lip positions, give a short drill for each problem sound.
Respond in %s.
Return ONLY a JSON object: {"summary": "2-3 sentence overall assessment", "tips": [{"sound": "short label: an IPA symbol or sound pair like \"θ/ð\" or \"ə\", at most 8 characters, no explanations here", "advice": "what is wrong and how to fix it"}]}`

func BuildFeedbackPrompt(text, heard, nativeLanguage string, issues []string) (system, user string) {
	var b strings.Builder
	fmt.Fprintf(&b, "TEXT:\n%s\n\nHEARD (IPA):\n%s\n", text, heard)
	if len(issues) > 0 {
		b.WriteString("\nDETECTED ISSUES:\n")
		for _, issue := range issues {
			b.WriteString("- " + issue + "\n")
		}
	}
	return fmt.Sprintf(feedbackSystem, nativeLanguage, nativeLanguage), b.String()
}

func ParseFeedback(raw string) (Feedback, error) {
	var out struct {
		Summary string `json:"summary"`
		Tips    []struct {
			Sound  string `json:"sound"`
			Advice string `json:"advice"`
		} `json:"tips"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Feedback{}, fmt.Errorf("parse speech feedback: %w", err)
	}
	fb := Feedback{Summary: strings.TrimSpace(out.Summary), Tips: make([]FeedbackTip, 0, len(out.Tips))}
	for _, t := range out.Tips {
		fb.Tips = append(fb.Tips, FeedbackTip{Sound: t.Sound, Advice: t.Advice})
	}
	return fb, nil
}
