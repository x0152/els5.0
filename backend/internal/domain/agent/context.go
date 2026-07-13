package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/els/backend/internal/domain/iam"
)

type ViewContext struct{}

func (ViewContext) Context(_ context.Context, rc RunContext) ([]LLMMessage, error) {
	v := rc.View
	if v == nil || (v.App == "" && v.Screen == "" && v.Title == "") {
		return nil, nil
	}
	var sb strings.Builder
	sb.WriteString("# Open right now\n")
	if v.App != "" {
		fmt.Fprintf(&sb, "- App: %s\n", v.App)
	}
	if v.Screen != "" {
		fmt.Fprintf(&sb, "- Screen: %s\n", v.Screen)
	}
	if v.Title != "" {
		fmt.Fprintf(&sb, "- Title: %s\n", v.Title)
	}
	writeSortedPairs(&sb, v.IDs)
	writeSortedPairs(&sb, v.State)
	if v.Info != "" {
		fmt.Fprintf(&sb, "- %s\n", v.Info)
	}
	return []LLMMessage{{Role: LLMRoleSystem, Content: sb.String()}}, nil
}

func writeSortedPairs(sb *strings.Builder, m map[string]string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if m[k] != "" {
			fmt.Fprintf(sb, "- %s: %s\n", k, m[k])
		}
	}
}

type actorKey struct{}

func WithActor(ctx context.Context, actor *iam.Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

func ActorFrom(ctx context.Context) *iam.Actor {
	actor, _ := ctx.Value(actorKey{}).(*iam.Actor)
	return actor
}
