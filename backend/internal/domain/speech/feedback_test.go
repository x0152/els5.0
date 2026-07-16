package speech_test

import (
	"strings"
	"testing"

	"github.com/els/backend/internal/domain/speech"
)

func TestBuildFeedbackPrompt(t *testing.T) {
	// arrange
	issues := []string{"think: expected θ, heard s"}

	// act
	system, user := speech.BuildFeedbackPrompt("I think", "aɪ s ɪ ŋ k", "Russian", issues)

	// assert
	if !strings.Contains(system, "Russian") {
		t.Fatalf("system prompt must mention native language: %s", system)
	}
	if !strings.Contains(user, "I think") || !strings.Contains(user, "aɪ s ɪ ŋ k") {
		t.Fatalf("user prompt must contain text and heard IPA: %s", user)
	}
	if !strings.Contains(user, "think: expected θ, heard s") {
		t.Fatalf("user prompt must list issues: %s", user)
	}
}

func TestParseFeedback(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
		tips    int
	}{
		{"valid", `{"summary":"Good job","tips":[{"sound":"θ","advice":"tongue between teeth"}]}`, false, 1},
		{"empty tips", `{"summary":"Perfect","tips":[]}`, false, 0},
		{"invalid json", `not json`, true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			fb, err := speech.ParseFeedback(tc.raw)

			// assert
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(fb.Tips) != tc.tips {
				t.Fatalf("expected %d tips, got %d", tc.tips, len(fb.Tips))
			}
		})
	}
}

func TestLookupPhoneme(t *testing.T) {
	// act
	info, ok := speech.LookupPhoneme("θ")

	// assert
	if !ok || info.Symbol != "θ" || info.Description == "" {
		t.Fatalf("expected guide entry for θ, got %+v (ok=%v)", info, ok)
	}
	if _, ok := speech.LookupPhoneme("zzz"); ok {
		t.Fatal("unexpected entry for unknown symbol")
	}
}
