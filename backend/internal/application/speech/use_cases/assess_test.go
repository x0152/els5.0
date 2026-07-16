package usecases_test

import (
	"context"
	"errors"
	"testing"

	usecases "github.com/els/backend/internal/application/speech/use_cases"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/speech"
)

func TestAssessUseCase(t *testing.T) {
	cases := []struct {
		name       string
		cmd        usecases.AssessCommand
		svcErr     error
		wantErr    error
		wantStrict float64
	}{
		{"missing audio", usecases.AssessCommand{Text: "hello", Strictness: 1}, nil, shared.ErrValidation, 0},
		{"missing text", usecases.AssessCommand{Audio: []byte{1}, Strictness: 1}, nil, shared.ErrValidation, 0},
		{"happy", usecases.AssessCommand{Audio: []byte{1}, Text: "hello", Strictness: 1.5}, nil, nil, 1.5},
		{"strictness out of range falls back", usecases.AssessCommand{Audio: []byte{1}, Text: "hello", Strictness: 99}, nil, nil, speech.DefaultStrictness},
		{"service failure", usecases.AssessCommand{Audio: []byte{1}, Text: "hello", Strictness: 1}, errors.New("boom"), nil, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			mock := &assessorMock{result: speech.Assessment{Overall: 90}, err: tc.svcErr}
			uc := usecases.NewAssessUseCase(mock)

			// act
			res, err := uc.Execute(context.Background(), nil, tc.cmd)

			// assert
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				if mock.calls != 0 {
					t.Fatal("assessor must not be called on validation error")
				}
				return
			}
			if tc.svcErr != nil {
				if err == nil {
					t.Fatal("expected service error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Overall != 90 {
				t.Fatalf("expected assessment passthrough, got %+v", res)
			}
			if mock.gotStrict != tc.wantStrict {
				t.Fatalf("expected strictness %v, got %v", tc.wantStrict, mock.gotStrict)
			}
		})
	}
}
