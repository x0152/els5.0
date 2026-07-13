package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultMaxIterations = 12

type Loop struct {
	llm           LLM
	context       []ContextPlugin
	tools         []ToolPlugin
	finish        []FinishPlugin
	maxIterations int
}

func NewLoop(llm LLM, ctxPlugins []ContextPlugin, toolPlugins []ToolPlugin, finishPlugins []FinishPlugin, maxIterations int) *Loop {
	if maxIterations <= 0 {
		maxIterations = defaultMaxIterations
	}
	return &Loop{llm: llm, context: ctxPlugins, tools: toolPlugins, finish: finishPlugins, maxIterations: maxIterations}
}

func (l *Loop) Run(ctx context.Context, rc RunContext) (<-chan Event, error) {
	if l.llm == nil {
		return nil, fmt.Errorf("agent loop: llm is nil")
	}

	// Phase 1: build context — each ContextPlugin adds system messages.
	messages := make([]LLMMessage, 0, len(rc.History)+8)
	for _, p := range l.context {
		extra, err := p.Context(ctx, rc)
		if err != nil {
			return nil, err
		}
		messages = append(messages, extra...)
	}
	for _, m := range rc.History {
		if llmMsg, ok := m.toLLM(); ok {
			messages = append(messages, llmMsg)
		}
	}
	if strings.TrimSpace(rc.UserMessage) != "" {
		messages = append(messages, LLMMessage{Role: LLMRoleUser, Content: rc.UserMessage})
	}

	// Phase 2: collect reason+act tools from all ToolPlugins.
	tools := []Tool{}
	for _, p := range l.tools {
		tools = append(tools, p.Tools(rc)...)
	}
	toolMap := make(map[string]Tool, len(tools))
	specs := make([]LLMToolSpec, 0, len(tools))
	for _, t := range tools {
		toolMap[t.Name] = t
		specs = append(specs, t.Spec())
	}

	out := make(chan Event, 32)
	go func() {
		defer close(out)
		l.iterate(ctx, rc, messages, specs, toolMap, out)
	}()
	return out, nil
}

func (l *Loop) iterate(ctx context.Context, rc RunContext, messages []LLMMessage, specs []LLMToolSpec, toolMap map[string]Tool, out chan<- Event) {
	model := rc.Model

	for iter := 0; iter < l.maxIterations; iter++ {
		if ctx.Err() != nil {
			out <- Event{Type: EventDone, FinishReason: FinishCanceled}
			return
		}

		stream, err := l.llm.ChatStream(ctx, LLMRequest{Model: model, Messages: messages, Tools: specs})
		if err != nil {
			out <- Event{Type: EventError, ErrMessage: err.Error()}
			return
		}

		var (
			reply     strings.Builder
			reasoning strings.Builder
			toolCalls []LLMToolCall
			usage     LLMUsage
			finish    string
		)
		for ev := range stream {
			switch ev.Type {
			case LLMEventText:
				if ev.TextDelta != "" {
					reply.WriteString(ev.TextDelta)
					out <- Event{Type: EventText, TextDelta: ev.TextDelta}
				}
			case LLMEventReasoning:
				reasoning.WriteString(ev.TextDelta)
			case LLMEventToolCalls:
				toolCalls = append(toolCalls, ev.ToolCalls...)
			case LLMEventFinish:
				usage = ev.Usage
				finish = ev.FinishReason
			case LLMEventError:
				out <- Event{Type: EventError, ErrMessage: ev.ErrMessage}
				return
			}
		}

		if len(toolCalls) == 0 {
			// Phase 3: post-process the final reply.
			for _, p := range l.finish {
				if err := p.Finish(ctx, FinishContext{Actor: rc.Actor, FinalText: reply.String()}); err != nil {
					out <- Event{Type: EventError, ErrMessage: err.Error()}
					return
				}
			}
			out <- Event{Type: EventAssistantTurn, Text: reply.String(), Usage: usage, Model: model, FinishReason: mapFinish(finish)}
			out <- Event{Type: EventDone, FinishReason: mapFinish(finish), Usage: usage, Model: model}
			return
		}

		messages = append(messages, LLMMessage{
			Role:             LLMRoleAssistant,
			Content:          reply.String(),
			ToolCalls:        toolCalls,
			ReasoningContent: reasoning.String(),
		})
		out <- Event{Type: EventAssistantTurn, Text: reply.String(), ToolCalls: toolCalls, Usage: usage, Model: model, FinishReason: FinishToolCalls}

		for _, call := range toolCalls {
			stepID := call.ID
			if stepID == "" {
				stepID = uuid.NewString()
			}
			tool, ok := toolMap[call.Name]
			label := call.Name
			if ok && tool.Label != nil {
				label = tool.Label(call.Arguments)
			}
			step := Step{ID: stepID, Tool: call.Name, Label: label, Args: call.Arguments, StartedAt: now()}
			if ok {
				step.Icon = tool.Icon
			}
			out <- Event{Type: EventToolStart, Step: &step}

			result := execute(ctx, tool, ok, call)
			step.Result = result
			step.FinishedAt = now()
			out <- Event{Type: EventToolEnd, StepID: stepID, ToolResult: result, Step: &step}

			messages = append(messages, LLMMessage{Role: LLMRoleTool, Content: result, ToolCallID: call.ID, ToolName: call.Name})
		}
	}

	out <- Event{Type: EventError, ErrMessage: fmt.Sprintf("max iterations reached: %d", l.maxIterations)}
}

func execute(ctx context.Context, t Tool, found bool, call LLMToolCall) string {
	if !found || t.Execute == nil {
		return fmt.Sprintf("error: unknown tool %q", call.Name)
	}
	res, err := t.Execute(ctx, call.Arguments)
	if err != nil {
		return "error: " + err.Error()
	}
	if strings.TrimSpace(res) == "" {
		return fmt.Sprintf("status: success (%s completed with no output)", call.Name)
	}
	return res
}

func mapFinish(s string) FinishReason {
	switch strings.TrimSpace(s) {
	case "tool_calls":
		return FinishToolCalls
	case "length":
		return FinishLength
	default:
		return FinishStop
	}
}

func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
