package lexicon

type UnitType string

const (
	UnitTypeSingle      UnitType = "single_unit"
	UnitTypeCompound    UnitType = "compound_unit"
	UnitTypePhrase      UnitType = "phrase_unit"
	UnitTypeEntity      UnitType = "entity_unit"
	UnitTypePhrasal     UnitType = "phrasal_unit"
	UnitTypeIdiom       UnitType = "idiom_unit"
	UnitTypeCollocation UnitType = "collocation_unit"
	UnitTypeProverb     UnitType = "proverb_unit"
	UnitTypeTime        UnitType = "time_unit"
	UnitTypeMoney       UnitType = "money_unit"
	UnitTypeQuantity    UnitType = "quantity_unit"
	UnitTypePercent     UnitType = "percent_unit"
	UnitTypeProduct     UnitType = "product_unit"
	UnitTypeTechnical   UnitType = "technical_unit"
)

type SegmentKind string

const (
	SegmentKindSentence SegmentKind = "sentence"
	SegmentKindSubtitle SegmentKind = "subtitle"
)

type Span struct {
	Position int
	SpanType string
	Start    int
	End      int
	Text     string
}

type BaseForm struct {
	Text     string
	POS      string
	IsStop   bool
	Language string
}

type Unit struct {
	UnitType    UnitType
	BaseForm    string
	POS         string
	SentenceIdx int
	Metadata    map[string]any
	Language    string
	Spans       []Span
}

type Cue struct {
	Index   int
	StartMs int
	EndMs   int
	Text    string
}

type Segment struct {
	Kind       SegmentKind
	SegmentIdx int
	StartPos   int
	EndPos     int
	Text       string
	Metadata   map[string]any
}

type Analysis struct {
	SentenceCount int
	Language      string
	Units         []Unit
	BaseForms     map[string]BaseForm
}

type Spot struct {
	Ref     int
	Example string
}

type MediaOccurrence struct {
	MediaID     string
	MediaType   string
	Title       string
	Kind        string
	SeriesTitle string
	Season      int
	Episode     int
	Author      string
	Count       int
	Spots       []Spot
}

type LemmaOccurrences struct {
	Lemma      string
	IsStop     bool
	Total      int
	MediaCount int
	Media      []MediaOccurrence
}
