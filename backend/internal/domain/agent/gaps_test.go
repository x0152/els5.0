package agent_test

import (
	"testing"

	"github.com/els/backend/internal/domain/agent"
)

func TestFillGap(t *testing.T) {
	cases := []struct {
		name    string
		content string
		ordinal int
		answer  string
		want    string
		found   bool
	}{
		{"first gap", "a {{x}} b {{y}}", 0, "one", "a {{x||one}} b {{y}}", true},
		{"second gap", "a {{x}} b {{y}}", 1, "two", "a {{x}} b {{y||two}}", true},
		{"replaces previous fill", "a {{x||old}}", 0, "new", "a {{x||new}}", true},
		{"sanitizes braces and newlines", "a {{x}}", 0, " {ne\nw} ", "a {{x||ne w}}", true},
		{"ordinal out of range", "a {{x}}", 3, "one", "a {{x}}", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			got, found := agent.FillGap(tc.content, tc.ordinal, tc.answer)
			// assert
			if got != tc.want || found != tc.found {
				t.Fatalf("got (%q, %v), want (%q, %v)", got, found, tc.want, tc.found)
			}
		})
	}
}
