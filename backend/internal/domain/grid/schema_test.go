package grid_test

import (
	"testing"

	"github.com/els/backend/internal/domain/grid"
)

func TestSchema_ColumnAndHasColumn(t *testing.T) {
	s := sampleSchema()
	col, ok := s.Column("email")
	if !ok || col.ID != "email" {
		t.Errorf("expected to find email column, got %v ok=%v", col, ok)
	}
	if s.HasColumn("missing") {
		t.Errorf("expected no missing column")
	}
}

func TestSchema_SourcesUniqueAndSorted(t *testing.T) {
	s := grid.Schema{
		Columns: []grid.Column{
			{ID: "owner", Type: grid.TypeRef, Ref: &grid.RefSpec{Source: grid.SourceID("zaccounts")}},
			{ID: "client", Type: grid.TypeRef, Ref: &grid.RefSpec{Source: grid.SourceID("aclients")}},
			{ID: "co_owner", Type: grid.TypeRef, Ref: &grid.RefSpec{Source: grid.SourceID("zaccounts")}},
		},
	}
	got := s.Sources()
	want := []grid.SourceID{"aclients", "zaccounts"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("at %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestHash_StableAndOrderInsensitive(t *testing.T) {
	a := []grid.Column{{ID: "name"}, {ID: "email"}}
	b := []grid.Column{{ID: "email"}, {ID: "name"}}
	if grid.Hash(a) != grid.Hash(b) {
		t.Errorf("hash must be order-insensitive")
	}
	if grid.Hash(a) == grid.Hash([]grid.Column{{ID: "name"}}) {
		t.Errorf("different columns should have different hashes")
	}
}
