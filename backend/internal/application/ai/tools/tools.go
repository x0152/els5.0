package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/agent"
	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/vocab"
)

type EventsReader interface {
	ListEvents(ctx context.Context, userID string) ([]core.Event, error)
}

type Plugin struct {
	tools []agent.Tool
}

func NewPlugin(reader EventsReader, words vocab.Repository) *Plugin {
	t := []agent.Tool{readRecentErrors(reader), currentTime()}
	t = append(t, vocabTools(words)...)
	return &Plugin{tools: t}
}

func (p *Plugin) Tools(_ agent.RunContext) []agent.Tool { return p.tools }

func readRecentErrors(reader EventsReader) agent.Tool {
	return agent.Tool{
		Name:        "read_recent_errors",
		Description: "Returns the current user's recent language mistakes from their learning events.",
		Icon:        "alert-triangle",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"limit": map[string]any{"type": "integer", "description": "How many mistakes to return (default 3)."},
			},
		},
		Label: func(string) string { return "Reading recent mistakes" },
		Execute: func(ctx context.Context, args string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			limit := 3
			if strings.TrimSpace(args) != "" {
				var a struct {
					Limit int `json:"limit"`
				}
				if json.Unmarshal([]byte(args), &a) == nil && a.Limit > 0 {
					limit = a.Limit
				}
			}
			events, err := reader.ListEvents(ctx, actor.AccountID().String())
			if err != nil {
				return "", err
			}
			out := make([]map[string]any, 0, limit)
			for _, e := range events {
				if e.Error == nil {
					continue
				}
				out = append(out, map[string]any{
					"occurred_at": e.OccurredAt.Format(time.RFC3339),
					"name":        e.Error.Name,
					"sentence":    e.Error.Sentence,
					"correction":  e.Error.Correction,
					"description": e.Error.Description,
				})
				if len(out) >= limit {
					break
				}
			}
			if len(out) == 0 {
				return "The user has no recorded mistakes yet.", nil
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return string(b), nil
		},
	}
}

func currentTime() agent.Tool {
	return agent.Tool{
		Name:        "current_time",
		Description: "Returns the current server date and time in UTC.",
		Icon:        "clock",
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Label:       func(string) string { return "Current time" },
		Execute: func(_ context.Context, _ string) (string, error) {
			return time.Now().UTC().Format(time.RFC3339), nil
		},
	}
}
