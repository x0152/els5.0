package api

import authx "github.com/els/backend/internal/utils/auth"

type ToolCallOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type MessageOutput struct {
	ID         string           `json:"id"`
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []ToolCallOutput `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolName   string           `json:"tool_name,omitempty"`
	Model      string           `json:"model,omitempty"`
	CreatedAt  string           `json:"created_at"`
}

type HistoryOutput struct {
	Model        string          `json:"model"`
	DefaultModel string          `json:"default_model"`
	Generating   bool            `json:"generating"`
	Messages     []MessageOutput `json:"messages"`
}

type ModelOutput struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by,omitempty"`
}

type ModelsOutput struct {
	Models   []ModelOutput `json:"models"`
	Selected string        `json:"selected"`
	Default  string        `json:"default"`
}

type OKOutput struct {
	OK bool `json:"ok"`
}

type HistoryInput struct {
	authx.BearerInput
}

type ModelsInput struct {
	authx.BearerInput
}

type SetModelInput struct {
	authx.BearerInput
	Body struct {
		Model string `json:"model"`
	}
}

type ResetInput struct {
	authx.BearerInput
}

type ClearInput struct {
	authx.BearerInput
}
