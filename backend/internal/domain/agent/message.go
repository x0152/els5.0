package agent

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleSeparator Role = "separator"
)

type Message struct {
	ID               string
	SessionID        string
	Role             Role
	Content          string
	ToolCalls        []LLMToolCall
	ToolCallID       string
	ToolName         string
	Model            string
	FinishReason     string
	ReasoningContent string
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
	CreatedAt        time.Time
}

func (m Message) toLLM() (LLMMessage, bool) {
	out := LLMMessage{Content: m.Content}
	switch m.Role {
	case RoleUser:
		out.Role = LLMRoleUser
	case RoleAssistant:
		out.Role = LLMRoleAssistant
		out.ToolCalls = m.ToolCalls
		out.ReasoningContent = m.ReasoningContent
	case RoleTool:
		out.Role = LLMRoleTool
		out.ToolCallID = m.ToolCallID
		out.ToolName = m.ToolName
	default:
		return LLMMessage{}, false
	}
	return out, true
}
