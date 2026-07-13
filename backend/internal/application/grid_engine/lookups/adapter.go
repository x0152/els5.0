package lookups

import (
	"context"

	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
)

type Item struct {
	Key   string
	Label string
}

type Page struct {
	Items      []Item
	NextCursor string
}

type MatchKind string

const (
	MatchByKey   MatchKind = "key"
	MatchByLabel MatchKind = "label"
)

type Resolution struct {
	Input     string
	Key       string
	Label     string
	MatchedBy MatchKind
}

type Adapter interface {
	Hydrate(ctx context.Context, actor *iam.Actor, keys []string) (map[string]string, error)
	Resolve(ctx context.Context, actor *iam.Actor, values []string) ([]Resolution, []string, error)
	Search(ctx context.Context, actor *iam.Actor, q string, limit int32, cursor string) (Page, error)
}

type Source struct {
	ID      grid.SourceID
	Adapter Adapter
}
