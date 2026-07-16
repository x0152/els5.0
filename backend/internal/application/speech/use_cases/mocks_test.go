package usecases_test

import (
	"context"

	"github.com/els/backend/internal/domain/speech"
)

type assessorMock struct {
	calls     int
	gotText   string
	gotStrict float64
	result    speech.Assessment
	err       error
}

func (m *assessorMock) Assess(_ context.Context, _ []byte, text string, strictness float64) (speech.Assessment, error) {
	m.calls++
	m.gotText = text
	m.gotStrict = strictness
	return m.result, m.err
}

type llmMock struct {
	available bool
	response  string
	err       error
	gotSystem string
	gotUser   string
}

func (m *llmMock) Available() bool { return m.available }

func (m *llmMock) Chat(_ context.Context, system, user string) (string, error) {
	m.gotSystem = system
	m.gotUser = user
	return m.response, m.err
}
