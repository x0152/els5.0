package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	usecases "github.com/els/backend/internal/application/speech/use_cases"
	"github.com/els/backend/internal/domain/shared"
)

func TestFeedbackUseCase(t *testing.T) {
	valid := usecases.FeedbackCommand{Text: "I think", Heard: "aɪ s ɪ ŋ k", NativeLanguage: "Russian"}

	cases := []struct {
		name    string
		cmd     usecases.FeedbackCommand
		llm     llmMock
		wantErr error
		wantAny bool
	}{
		{"missing text", usecases.FeedbackCommand{Heard: "x"}, llmMock{available: true}, shared.ErrValidation, true},
		{"missing heard", usecases.FeedbackCommand{Text: "x"}, llmMock{available: true}, shared.ErrValidation, true},
		{"llm unavailable", valid, llmMock{available: false}, shared.ErrValidation, true},
		{"llm failure", valid, llmMock{available: true, err: errors.New("boom")}, nil, true},
		{"bad llm json", valid, llmMock{available: true, response: "oops"}, nil, true},
		{"happy", valid, llmMock{available: true, response: `{"summary":"ok","tips":[{"sound":"θ","advice":"do"}]}`}, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			uc := usecases.NewFeedbackUseCase(&tc.llm)

			// act
			fb, err := uc.Execute(context.Background(), nil, tc.cmd)

			// assert
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}
			if tc.wantAny {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fb.Summary != "ok" || len(fb.Tips) != 1 {
				t.Fatalf("unexpected feedback: %+v", fb)
			}
			if !strings.Contains(tc.llm.gotSystem, "Russian") {
				t.Fatal("native language must reach the prompt")
			}
		})
	}
}
