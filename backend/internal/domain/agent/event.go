package agent

type EventType string

const (
	EventText          EventType = "text"
	EventAssistantTurn EventType = "assistant_turn"
	EventToolStart     EventType = "tool_start"
	EventToolEnd       EventType = "tool_end"
	EventError         EventType = "error"
	EventDone          EventType = "done"
)

type FinishReason string

const (
	FinishStop      FinishReason = "stop"
	FinishToolCalls FinishReason = "tool_calls"
	FinishLength    FinishReason = "length"
	FinishCanceled  FinishReason = "canceled"
)

type Step struct {
	ID         string
	Tool       string
	Label      string
	Icon       string
	Args       string
	Result     string
	StartedAt  string
	FinishedAt string
}

type Event struct {
	Type         EventType
	TextDelta    string
	Text         string
	ToolCalls    []LLMToolCall
	Step         *Step
	StepID       string
	ToolResult   string
	Usage        LLMUsage
	Model        string
	FinishReason FinishReason
	ErrMessage   string
}
