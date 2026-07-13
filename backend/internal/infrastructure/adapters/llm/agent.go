package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"github.com/els/backend/internal/domain/agent"
	"github.com/els/backend/internal/domain/shared/ports"
)

type AgentConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

type AgentClient struct {
	cfg      AgentConfig
	resolver ports.AIProviderResolver
}

func NewAgentClient(cfg AgentConfig) *AgentClient {
	return &AgentClient{cfg: cfg}
}

func NewAgentClientWithResolver(cfg AgentConfig, resolver ports.AIProviderResolver) *AgentClient {
	return &AgentClient{cfg: cfg, resolver: resolver}
}

func (c *AgentClient) resolve(ctx context.Context) (openai.Client, string, bool) {
	cfg := c.cfg
	if c.resolver != nil {
		if r := c.resolver.Resolve(ctx); !r.IsEmpty() {
			cfg = AgentConfig{BaseURL: r.BaseURL, APIKey: r.APIKey, Model: r.Model}
		}
	}
	opts := []option.RequestOption{}
	if cfg.APIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.APIKey))
	}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(cfg.BaseURL, "/")))
	}
	reasoning := strings.Contains(strings.ToLower(cfg.BaseURL+cfg.Model), "deepseek")
	return openai.NewClient(opts...), strings.TrimSpace(cfg.Model), reasoning
}

func (c *AgentClient) DefaultModel() string {
	_, model, _ := c.resolve(context.Background())
	return model
}

func (c *AgentClient) ListModels(ctx context.Context) ([]agent.LLMModel, error) {
	client, _, _ := c.resolve(ctx)
	page, err := client.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("llm: list models: %w", err)
	}
	out := make([]agent.LLMModel, 0, len(page.Data))
	for _, m := range page.Data {
		out = append(out, agent.LLMModel{ID: m.ID, OwnedBy: m.OwnedBy})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (c *AgentClient) ChatStream(ctx context.Context, req agent.LLMRequest) (<-chan agent.LLMEvent, error) {
	client, defaultModel, reasoning := c.resolve(ctx)
	model := req.Model
	if model == "" {
		model = defaultModel
	}
	if model == "" {
		return nil, fmt.Errorf("llm: model not configured")
	}

	messages, err := c.toOpenAIMessages(req.Messages, reasoning)
	if err != nil {
		return nil, err
	}
	params := openai.ChatCompletionNewParams{
		Model:         shared.ChatModel(model),
		Messages:      messages,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)},
	}
	if tools := toOpenAITools(req.Tools); len(tools) > 0 {
		params.Tools = tools
	}

	stream := client.Chat.Completions.NewStreaming(ctx, params)
	out := make(chan agent.LLMEvent, 32)
	go func() {
		defer close(out)
		defer stream.Close()

		accum := make(map[int64]*toolCallAccum)
		order := make([]int64, 0)
		var finishReason string
		var usage agent.LLMUsage

		for stream.Next() {
			chunk := stream.Current()
			for _, choice := range chunk.Choices {
				if r := reasoningDelta(choice.Delta); r != "" {
					out <- agent.LLMEvent{Type: agent.LLMEventReasoning, TextDelta: r}
				}
				if d := choice.Delta.Content; d != "" {
					out <- agent.LLMEvent{Type: agent.LLMEventText, TextDelta: d}
				}
				for _, tc := range choice.Delta.ToolCalls {
					a, ok := accum[tc.Index]
					if !ok {
						a = &toolCallAccum{}
						accum[tc.Index] = a
						order = append(order, tc.Index)
					}
					if tc.ID != "" {
						a.ID = tc.ID
					}
					a.Name += tc.Function.Name
					a.Arguments += tc.Function.Arguments
				}
				if choice.FinishReason != "" {
					finishReason = choice.FinishReason
				}
			}
			if chunk.Usage.TotalTokens > 0 || chunk.Usage.PromptTokens > 0 {
				usage = agent.LLMUsage{
					PromptTokens:     tokenCount(chunk.Usage.PromptTokens),
					CompletionTokens: tokenCount(chunk.Usage.CompletionTokens),
					TotalTokens:      tokenCount(chunk.Usage.TotalTokens),
				}
			}
		}
		if err := stream.Err(); err != nil {
			out <- agent.LLMEvent{Type: agent.LLMEventError, ErrMessage: err.Error()}
			return
		}

		if len(order) > 0 {
			sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
			calls := make([]agent.LLMToolCall, 0, len(order))
			for _, idx := range order {
				a := accum[idx]
				calls = append(calls, agent.LLMToolCall{ID: a.ID, Name: a.Name, Arguments: a.Arguments})
			}
			out <- agent.LLMEvent{Type: agent.LLMEventToolCalls, ToolCalls: calls}
		}
		out <- agent.LLMEvent{Type: agent.LLMEventFinish, FinishReason: finishReason, Usage: usage}
	}()
	return out, nil
}

func tokenCount(n int64) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(n) // #nosec G115 -- clamped to int32 range above.
}

type toolCallAccum struct {
	ID        string
	Name      string
	Arguments string
}

func (c *AgentClient) toOpenAIMessages(in []agent.LLMMessage, reasoningPassthrough bool) ([]openai.ChatCompletionMessageParamUnion, error) {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(in))
	for _, m := range in {
		switch m.Role {
		case agent.LLMRoleSystem:
			out = append(out, openai.SystemMessage(m.Content))
		case agent.LLMRoleUser:
			out = append(out, openai.UserMessage(m.Content))
		case agent.LLMRoleAssistant:
			assistant := openai.ChatCompletionAssistantMessageParam{}
			if m.Content != "" {
				assistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.String(m.Content)}
			}
			if len(m.ToolCalls) > 0 {
				calls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(m.ToolCalls))
				for _, tc := range m.ToolCalls {
					calls = append(calls, openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID:       tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{Name: tc.Name, Arguments: tc.Arguments},
						},
					})
				}
				assistant.ToolCalls = calls
				if m.ReasoningContent != "" || reasoningPassthrough {
					assistant.SetExtraFields(map[string]any{"reasoning_content": m.ReasoningContent})
				}
			}
			out = append(out, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant})
		case agent.LLMRoleTool:
			out = append(out, openai.ToolMessage(m.Content, m.ToolCallID))
		default:
			return nil, fmt.Errorf("llm: unsupported role %q", m.Role)
		}
	}
	return out, nil
}

func reasoningDelta(delta openai.ChatCompletionChunkChoiceDelta) string {
	field, ok := delta.JSON.ExtraFields["reasoning_content"]
	if !ok || !field.Valid() {
		return ""
	}
	var s string
	if err := json.Unmarshal([]byte(field.Raw()), &s); err != nil {
		return ""
	}
	return s
}

func toOpenAITools(specs []agent.LLMToolSpec) []openai.ChatCompletionToolUnionParam {
	if len(specs) == 0 {
		return nil
	}
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(specs))
	for _, s := range specs {
		params := shared.FunctionParameters{}
		for k, v := range s.Parameters {
			params[k] = v
		}
		out = append(out, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: shared.FunctionDefinitionParam{
					Name:        s.Name,
					Description: openai.String(s.Description),
					Parameters:  params,
				},
			},
		})
	}
	return out
}
