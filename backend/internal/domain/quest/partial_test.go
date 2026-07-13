package quest_test

import (
	"testing"

	"github.com/els/backend/internal/domain/quest"
)

func TestParsePartialWorld(t *testing.T) {
	cases := []struct {
		name          string
		raw           string
		narration     string
		narrationDone bool
		responses     []quest.PartialLine
		nilResult     bool
	}{
		{
			name:      "empty stream",
			raw:       "",
			nilResult: true,
		},
		{
			name:      "no json started",
			raw:       "```json",
			nilResult: true,
		},
		{
			name:      "narration cut mid word",
			raw:       `{"narration":"Ray freezes and sta`,
			narration: "Ray freezes and sta",
		},
		{
			name:          "narration complete no responses yet",
			raw:           `{"narration":"Ray freezes.","narrationVoice":"Jasper","responses":[`,
			narration:     "Ray freezes.",
			narrationDone: true,
		},
		{
			name:          "first response cut mid text",
			raw:           `{"narration":"Ray freezes.","responses":[{"name":"Ray Morrison","voice":"Jasper","text":"That is quite a the`,
			narration:     "Ray freezes.",
			narrationDone: true,
			responses:     []quest.PartialLine{{Name: "Ray Morrison", Text: "That is quite a the"}},
		},
		{
			name:          "two responses second cut",
			raw:           `{"narration":"","responses":[{"name":"Ray","voice":"Jasper","text":"Hello."},{"name":"Dana","voice":"Bella","text":"I saw`,
			narrationDone: true,
			responses: []quest.PartialLine{
				{Name: "Ray", Text: "Hello.", Done: true},
				{Name: "Dana", Text: "I saw"},
			},
		},
		{
			name:      "escaped quotes inside text",
			raw:       `{"narration":"He says \"wait\" and`,
			narration: `He says "wait" and`,
		},
		{
			name:      "trailing backslash cut",
			raw:       `{"narration":"Line one\`,
			narration: "Line one",
		},
		{
			name:      "markdown fence prefix",
			raw:       "```json\n" + `{"narration":"The shop is quiet.`,
			narration: "The shop is quiet.",
		},
		{
			name:      "name present text missing",
			raw:       `{"responses":[{"name":"Ray","voice":`,
			responses: []quest.PartialLine{{Name: "Ray"}},
		},
		{
			name:      "unclosed name skipped",
			raw:       `{"responses":[{"name":"Ray Morri`,
			nilResult: true,
		},
		{
			name:          "fully closed response",
			raw:           `{"narration":"Quiet.","responses":[{"name":"Ray","voice":"Jasper","text":"Hello there."}]`,
			narration:     "Quiet.",
			narrationDone: true,
			responses:     []quest.PartialLine{{Name: "Ray", Text: "Hello there.", Done: true}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			got := quest.ParsePartialWorld(tc.raw)

			// assert
			if tc.nilResult {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected result, got nil")
			}
			if got.Narration != tc.narration {
				t.Fatalf("narration: expected %q, got %q", tc.narration, got.Narration)
			}
			if got.NarrationDone != tc.narrationDone {
				t.Fatalf("narrationDone: expected %v, got %v", tc.narrationDone, got.NarrationDone)
			}
			if len(got.Responses) != len(tc.responses) {
				t.Fatalf("responses: expected %d, got %d (%+v)", len(tc.responses), len(got.Responses), got.Responses)
			}
			for i := range tc.responses {
				if got.Responses[i] != tc.responses[i] {
					t.Fatalf("response %d: expected %+v, got %+v", i, tc.responses[i], got.Responses[i])
				}
			}
		})
	}
}
