package grid

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type Schema struct {
	Columns []Column
	Version string
}

func (s Schema) Column(id ColumnID) (Column, bool) {
	for _, c := range s.Columns {
		if c.ID == id {
			return c, true
		}
	}
	return Column{}, false
}

func (s Schema) HasColumn(id ColumnID) bool {
	_, ok := s.Column(id)
	return ok
}

func (s Schema) Sources() []SourceID {
	seen := map[SourceID]struct{}{}
	out := []SourceID{}
	for _, c := range s.Columns {
		if !c.IsRef() {
			continue
		}
		src := c.Ref.Source
		if _, ok := seen[src]; ok {
			continue
		}
		seen[src] = struct{}{}
		out = append(out, src)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func Hash(columns []Column) string {
	cp := make([]Column, len(columns))
	copy(cp, columns)
	sort.Slice(cp, func(i, j int) bool { return cp[i].ID < cp[j].ID })
	b, _ := json.Marshal(cp)
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
