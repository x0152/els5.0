package usecases

import (
	"time"

	"github.com/els/backend/internal/application/grid_engine/lookups"
	"github.com/els/backend/internal/domain/grid"
)

type DescribeResult struct {
	SchemaVersion string
	Columns       []grid.Column
	Sources       []grid.SourceID
	Rows          []grid.Row
	Total         int64
	Limit         int32
	Offset        int32
	RefsHydrated  map[grid.SourceID]map[string]string
	GeneratedAt   time.Time
}

type ApplyResult struct {
	SchemaVersion string
	Applied       []grid.OpResult
	Failed        []grid.OpError
}

type LookupQueryRequest struct {
	Source grid.SourceID
	Values []string
	Q      string
	Limit  int32
	Cursor string
}

type LookupQueryResult struct {
	Source      grid.SourceID
	Resolutions []lookups.Resolution
	Unresolved  []string
	Items       []lookups.Item
	NextCursor  string
}

type LookupResult struct {
	Queries []LookupQueryResult
}
