package api

import "time"

type ColumnDTO struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Type        string          `json:"type"`
	Required    bool            `json:"required,omitempty"`
	Readonly    bool            `json:"readonly,omitempty"`
	Enum        []EnumOptionDTO `json:"enum,omitempty"`
	Ref         *RefSpecDTO     `json:"ref,omitempty"`
	Constraints *ConstraintsDTO `json:"constraints,omitempty"`
}

type EnumOptionDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type RefSpecDTO struct {
	Source     string `json:"source"`
	KeyField   string `json:"key_field"`
	LabelField string `json:"label_field"`
	Multi      bool   `json:"multi,omitempty"`
}

type ConstraintsDTO struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Unique    bool     `json:"unique,omitempty"`
}

type RowDTO struct {
	ID          string         `json:"id"`
	BaseVersion int64          `json:"base_version"`
	Cells       map[string]any `json:"cells"`
}

type RefsHydratedDTO map[string]map[string]string

type DescribeGridInput struct {
	Authorization string `header:"Authorization" doc:"Bearer <token>"`
	Limit         int32  `query:"limit" doc:"page size" default:"50"`
	Offset        int32  `query:"offset" doc:"page offset" default:"0"`
}

func (in DescribeGridInput) GetAuthorization() string { return in.Authorization }

type DescribeGridOutput struct {
	SchemaVersion string          `json:"schema_version"`
	Columns       []ColumnDTO     `json:"columns"`
	Sources       []string        `json:"sources"`
	Rows          []RowDTO        `json:"rows"`
	Total         int64           `json:"total"`
	Limit         int32           `json:"limit"`
	Offset        int32           `json:"offset"`
	RefsHydrated  RefsHydratedDTO `json:"refs_hydrated"`
	GeneratedAt   time.Time       `json:"generated_at"`
}

type OpDTO struct {
	Kind        string         `json:"kind" enum:"create,update,delete"`
	TempID      string         `json:"temp_id,omitempty"`
	ID          string         `json:"id,omitempty"`
	BaseVersion int64          `json:"base_version,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

type ApplyGridInput struct {
	Authorization string `header:"Authorization" doc:"Bearer <token>"`
	Body          struct {
		SchemaVersion string  `json:"schema_version" doc:"schema version hash received from GET /grid"`
		Operations    []OpDTO `json:"operations"`
	}
}

func (in ApplyGridInput) GetAuthorization() string { return in.Authorization }

type OpResultDTO struct {
	Index       int    `json:"index"`
	Kind        string `json:"kind"`
	TempID      string `json:"temp_id,omitempty"`
	ID          string `json:"id"`
	BaseVersion int64  `json:"base_version"`
}

type OpErrorDTO struct {
	Index   int    `json:"index"`
	TempID  string `json:"temp_id,omitempty"`
	ID      string `json:"id,omitempty"`
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ApplyGridOutput struct {
	SchemaVersion string        `json:"schema_version"`
	Applied       []OpResultDTO `json:"applied"`
	Failed        []OpErrorDTO  `json:"failed"`
}

type LookupQueryDTO struct {
	Source string   `json:"source"`
	Values []string `json:"values,omitempty"`
	Q      string   `json:"q,omitempty"`
	Limit  int32    `json:"limit,omitempty"`
	Cursor string   `json:"cursor,omitempty"`
}

type LookupGridInput struct {
	Authorization string `header:"Authorization" doc:"Bearer <token>"`
	Body          struct {
		Queries []LookupQueryDTO `json:"queries"`
	}
}

func (in LookupGridInput) GetAuthorization() string { return in.Authorization }

type LookupItemDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type LookupResolutionDTO struct {
	Input     string `json:"input"`
	Key       string `json:"key,omitempty"`
	Label     string `json:"label,omitempty"`
	MatchedBy string `json:"matched_by,omitempty"`
	Resolved  bool   `json:"resolved"`
}

type LookupQueryResultDTO struct {
	Source      string                `json:"source"`
	Resolutions []LookupResolutionDTO `json:"resolutions,omitempty"`
	Items       []LookupItemDTO       `json:"items,omitempty"`
	NextCursor  string                `json:"next_cursor,omitempty"`
}

type LookupGridOutput struct {
	Queries []LookupQueryResultDTO `json:"queries"`
}
