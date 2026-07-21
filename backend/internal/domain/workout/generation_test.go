package workout_test

import (
	"strings"
	"testing"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/workout"
)

func cues(texts ...string) []films.Cue {
	out := []films.Cue{}
	for i, t := range texts {
		out = append(out, films.Cue{Index: i + 1, StartMs: i * 3000, EndMs: i*3000 + 2500, Text: t})
	}
	return out
}

func TestParseSegments(t *testing.T) {
	subtitle := cues("Hello there", "How are you doing today my friend", "Fine")

	t.Run("maps cue ranges to milliseconds", func(t *testing.T) {
		// arrange
		raw := `{"segments":[{"from_cue":1,"to_cue":3,"recap":"","summary":"An opening chat.","phrases":[{"cue":2,"text":"How are you doing","level":"A2"}]}]}`
		// act
		segments, err := workout.ParseSegments(raw, subtitle)
		// assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if segments[0].StartMs != 0 || segments[0].EndMs != 8500 {
			t.Fatalf("range wrong: %+v", segments[0])
		}
		if segments[0].Phrases[0].StartMs != 3000 {
			t.Fatalf("phrase cue not resolved: %+v", segments[0].Phrases[0])
		}
	})

	t.Run("drops segments with unknown cues", func(t *testing.T) {
		// arrange
		raw := `{"segments":[{"from_cue":1,"to_cue":99,"summary":"bad"}]}`
		// act
		_, err := workout.ParseSegments(raw, subtitle)
		// assert
		if err == nil {
			t.Fatal("expected error for no valid segments")
		}
	})
}

func TestParseQuestions(t *testing.T) {
	// arrange
	raw := `{"questions":[
		{"text":"Who spoke first?","options":["A","B","C","D"],"answer":1},
		{"text":"","options":["A","B","C","D"],"answer":0},
		{"text":"Bad options","options":["A"],"answer":0}
	]}`
	// act
	questions, err := workout.ParseQuestions(raw)
	// assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 1 || questions[0].Answer != 1 {
		t.Fatalf("filtering wrong: %+v", questions)
	}
}

func TestBuildGrammarPrompt(t *testing.T) {
	t.Run("uses the learner's mistakes when present", func(t *testing.T) {
		// arrange
		focuses := []workout.GrammarFocus{{Rule: "past simple", Sentence: "I go there yesterday"}}
		// act
		_, user := workout.BuildGrammarPrompt(focuses, "a road trip", "B1", "- blocks -")
		// assert
		if !strings.Contains(user, "past simple") || !strings.Contains(user, "I go there yesterday") {
			t.Fatalf("mistakes missing from prompt: %s", user)
		}
	})

	t.Run("falls back to level-typical grammar without mistakes", func(t *testing.T) {
		// act
		_, user := workout.BuildGrammarPrompt(nil, "a road trip", "B1", "- blocks -")
		// assert
		if !strings.Contains(user, "no recorded mistakes") || !strings.Contains(user, "a road trip") {
			t.Fatalf("fallback prompt wrong: %s", user)
		}
	})
}

func TestDictationLines(t *testing.T) {
	// arrange
	subtitle := cues(
		"Yes",
		"I really think we should go there together tomorrow",
		"[door slams]",
		"You never told me what happened after the party last night",
	)
	// act
	lines := workout.DictationLines(subtitle, "f1", 5)
	// assert
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2 typable sentences: %+v", len(lines), lines)
	}
	for _, l := range lines {
		if l.FilmID != "f1" || l.EndMs == 0 {
			t.Fatalf("cue reference lost: %+v", l)
		}
	}
}
