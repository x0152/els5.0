package grid

type Grid[E any] struct {
	schema  Schema
	row     func(E) Row
	apply   func(E, map[ColumnID]any) error
	sources []SourceID
}

type Spec[E any] struct {
	Columns    []Column
	Row        func(E) Row
	ApplyPatch func(E, map[ColumnID]any) error
	Sources    []SourceID
}

func New[E any](s Spec[E]) Grid[E] {
	schema := Schema{Columns: s.Columns, Version: Hash(s.Columns)}
	sources := s.Sources
	if sources == nil {
		sources = schema.Sources()
	}
	return Grid[E]{
		schema:  schema,
		row:     s.Row,
		apply:   s.ApplyPatch,
		sources: sources,
	}
}

func (g Grid[E]) Schema() Schema      { return g.schema }
func (g Grid[E]) Version() string     { return g.schema.Version }
func (g Grid[E]) Sources() []SourceID { return g.sources }

func (g Grid[E]) HasSource(s SourceID) bool {
	for _, x := range g.sources {
		if x == s {
			return true
		}
	}
	return false
}

func (g Grid[E]) RowOf(e E) Row { return g.row(e) }

func (g Grid[E]) ApplyPatch(e E, data map[ColumnID]any) error {
	return g.apply(e, data)
}
