package agent

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
)

type RunContext struct {
	Actor       *iam.Actor
	History     []Message
	UserMessage string
	Model       string
	View        *View
}

type View struct {
	App    string
	Screen string
	Title  string
	Info   string
	IDs    map[string]string
	State  map[string]string
}

type FinishContext struct {
	Actor     *iam.Actor
	FinalText string
}

type ContextPlugin interface {
	Context(ctx context.Context, rc RunContext) ([]LLMMessage, error)
}

type ToolPlugin interface {
	Tools(rc RunContext) []Tool
}

type FinishPlugin interface {
	Finish(ctx context.Context, fc FinishContext) error
}
