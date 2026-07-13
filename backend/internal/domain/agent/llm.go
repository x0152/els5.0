package agent

import "context"

type LLMRole string

const (
	LLMRoleSystem    LLMRole = "system"
	LLMRoleUser      LLMRole = "user"
	LLMRoleAssistant LLMRole = "assistant"
	LLMRoleTool      LLMRole = "tool"
)

type LLMToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type LLMMessage struct {
	Role             LLMRole
	Content          string
	ToolCalls        []LLMToolCall
	ToolCallID       string
	ToolName         string
	ReasoningContent string
}

type LLMToolSpec struct {
	Name        string
	Description string
	Parameters  map[string]any
}

type LLMUsage struct {
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
}

type LLMEventType string

const (
	LLMEventText      LLMEventType = "text"
	LLMEventReasoning LLMEventType = "reasoning"
	LLMEventToolCalls LLMEventType = "tool_calls"
	LLMEventFinish    LLMEventType = "finish"
	LLMEventError     LLMEventType = "error"
)

type LLMEvent struct {
	Type         LLMEventType
	TextDelta    string
	ToolCalls    []LLMToolCall
	FinishReason string
	Usage        LLMUsage
	ErrMessage   string
}

type LLMRequest struct {
	Model    string
	Messages []LLMMessage
	Tools    []LLMToolSpec
}

type LLMModel struct {
	ID      string
	OwnedBy string
}

type LLM interface {
	ChatStream(ctx context.Context, req LLMRequest) (<-chan LLMEvent, error)
	ListModels(ctx context.Context) ([]LLMModel, error)
	DefaultModel() string
}
