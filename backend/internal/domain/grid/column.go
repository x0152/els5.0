package grid

type ColumnID string

type SourceID string

type Type string

const (
	TypeText     Type = "text"
	TypeEmail    Type = "email"
	TypeInt      Type = "int"
	TypeFloat    Type = "float"
	TypeBool     Type = "bool"
	TypeDate     Type = "date"
	TypeDateTime Type = "datetime"
	TypeEnum     Type = "enum"
	TypeRef      Type = "ref"
)

type EnumOption struct {
	Value string
	Label string
}

type RefSpec struct {
	Source     SourceID
	KeyField   string
	LabelField string
	Multi      bool
}

type Constraints struct {
	MinLength *int
	MaxLength *int
	Min       *float64
	Max       *float64
	Pattern   string
	Unique    bool
}

type Column struct {
	ID          ColumnID
	Title       string
	Type        Type
	Required    bool
	Readonly    bool
	Enum        []EnumOption
	Ref         *RefSpec
	Constraints *Constraints
}

func (c Column) IsRef() bool { return c.Type == TypeRef && c.Ref != nil }
