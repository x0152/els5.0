package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/practice"
)

type Checker interface {
	Check(ctx context.Context, theory, instruction, answer string) (practice.CheckResult, error)
}

type CheckFreeUseCase struct {
	sources SourceProvider
	checker Checker
}

func NewCheckFreeUseCase(sources SourceProvider, checker Checker) *CheckFreeUseCase {
	return &CheckFreeUseCase{sources: sources, checker: checker}
}

type CheckFreeCommand struct {
	Kind        practice.Kind
	Number      int
	Instruction string
	Answer      string
}

func (uc *CheckFreeUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd CheckFreeCommand) (practice.CheckResult, error) {
	// 1. Pull chapter theory as context for checking (optional).
	var theory string
	if src, err := uc.sources.Source(ctx, cmd.Kind, cmd.Number); err == nil {
		theory = src.Theory
	}
	// 2. Send the free-form answer to the LLM for checking.
	return uc.checker.Check(ctx, theory, cmd.Instruction, cmd.Answer)
}
