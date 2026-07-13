package api

import (
	"time"

	usecases "github.com/els/backend/internal/application/ai/use_cases"
	"github.com/els/backend/internal/domain/agent"
)

func toHistoryOutput(res usecases.HistoryResult) HistoryOutput {
	out := HistoryOutput{Model: res.Model, DefaultModel: res.DefaultModel, Generating: res.Generating, Messages: []MessageOutput{}}
	for _, m := range res.Messages {
		out.Messages = append(out.Messages, toMessageOutput(m))
	}
	return out
}

func toMessageOutput(m agent.Message) MessageOutput {
	out := MessageOutput{
		ID:         m.ID,
		Role:       string(m.Role),
		Content:    m.Content,
		ToolCallID: m.ToolCallID,
		ToolName:   m.ToolName,
		Model:      m.Model,
		CreatedAt:  m.CreatedAt.UTC().Format(time.RFC3339),
	}
	for _, c := range m.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, ToolCallOutput{ID: c.ID, Name: c.Name, Arguments: c.Arguments})
	}
	return out
}

func toModelsOutput(res usecases.ModelsResult) ModelsOutput {
	out := ModelsOutput{Selected: res.Selected, Default: res.Default, Models: []ModelOutput{}}
	for _, m := range res.Models {
		out.Models = append(out.Models, ModelOutput{ID: m.ID, OwnedBy: m.OwnedBy})
	}
	return out
}
