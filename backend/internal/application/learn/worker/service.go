package worker

import (
	"context"
	"fmt"

	"github.com/els/backend/internal/domain/book"
	"github.com/els/backend/internal/domain/practice"
)

type LLM interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type Service struct {
	llm LLM
}

func NewService(llm LLM) *Service { return &Service{llm: llm} }

func (s *Service) Available() bool { return s.llm != nil && s.llm.Available() }

func (s *Service) Plan(ctx context.Context, src practice.Source) (string, []practice.PlanItem, error) {
	if !s.Available() {
		return "", nil, fmt.Errorf("generation is not available")
	}
	system, user := practice.BuildPlanPrompt(src)
	raw, err := s.llm.Chat(ctx, system, user)
	if err != nil {
		return "", nil, err
	}
	return practice.ParsePlan(raw)
}

func (s *Service) GenerateExercise(ctx context.Context, src practice.Source, item practice.PlanItem, number int) (string, error) {
	system, user := practice.BuildExercisePrompt(src, item, number)
	raw, err := s.llm.Chat(ctx, system, user)
	if err != nil {
		return "", err
	}
	return practice.ParseGeneratedExercise(raw)
}

func (s *Service) GenerateChapter(ctx context.Context, bookSlug, topic string) (book.Chapter, error) {
	if !s.Available() {
		return book.Chapter{}, fmt.Errorf("generation is not available")
	}
	system, user := book.BuildChapterPrompt(bookSlug, topic)
	raw, err := s.llm.Chat(ctx, system, user)
	if err != nil {
		return book.Chapter{}, err
	}
	return book.ParseGeneratedChapter(bookSlug, raw)
}

func (s *Service) Check(ctx context.Context, theory, instruction, answer string) (practice.CheckResult, error) {
	if !s.Available() {
		return practice.CheckResult{}, fmt.Errorf("checking is not available")
	}
	system, user := practice.BuildCheckPrompt(theory, instruction, answer)
	raw, err := s.llm.Chat(ctx, system, user)
	if err != nil {
		return practice.CheckResult{}, err
	}
	return practice.ParseCheckResult(raw)
}
