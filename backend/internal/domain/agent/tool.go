package agent

import "context"

type Tool struct {
	Name        string
	Description string
	Icon        string
	Parameters  map[string]any
	Label       func(args string) string
	Execute     func(ctx context.Context, args string) (string, error)
}

func (t Tool) Spec() LLMToolSpec {
	return LLMToolSpec{Name: t.Name, Description: t.Description, Parameters: t.Parameters}
}
