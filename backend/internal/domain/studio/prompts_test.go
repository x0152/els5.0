package studio_test

import (
	"errors"
	"testing"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/studio"
)

func TestParseEnrichment(t *testing.T) {
	// arrange
	raw := `{"transcription":"həˈləʊ","translation":"привет","explanation":"A greeting.","explanation_native":"Приветствие.","example":"Hello there!"}`

	// act
	got, err := studio.ParseEnrichment(raw)

	// assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Transcription != "həˈləʊ" || got.Translation != "привет" || got.Example != "Hello there!" {
		t.Errorf("got %+v", got)
	}
	if got.Explanation != "A greeting." || got.ExplanationNative != "Приветствие." {
		t.Errorf("got %+v", got)
	}
}

func TestParseTaskRejectsEmpty(t *testing.T) {
	// act
	_, err := studio.ParseTask(`{"task":"  "}`)

	// assert
	if err == nil {
		t.Fatal("expected error for empty task")
	}
}

func TestMarkSkill(t *testing.T) {
	cases := []struct {
		skill   string
		wantErr bool
	}{
		{studio.SkillListened, false},
		{studio.SkillSpoken, false},
		{studio.SkillWritten, false},
		{studio.SkillRecalled, false},
		{"reading", true},
	}
	for _, c := range cases {
		t.Run(c.skill, func(t *testing.T) {
			// arrange
			item := studio.Item{Text: "hi", AccountID: "a"}

			// act
			err := item.MarkSkill(c.skill)

			// assert
			if c.wantErr {
				if !errors.Is(err, shared.ErrValidation) {
					t.Errorf("want validation error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDoneRequiresAllSkills(t *testing.T) {
	// arrange
	item := studio.Item{Listened: true, Spoken: true, Written: true}

	// assert
	if item.Done() {
		t.Error("done without recalled")
	}

	// act
	item.Recalled = true

	// assert
	if !item.Done() {
		t.Error("all skills marked, want done")
	}
}
