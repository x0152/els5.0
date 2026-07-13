package grid

type Row struct {
	ID          string
	BaseVersion int64
	Cells       map[ColumnID]any
}

type OpKind string

const (
	OpCreate OpKind = "create"
	OpUpdate OpKind = "update"
	OpDelete OpKind = "delete"
)

func (k OpKind) IsValid() bool {
	switch k {
	case OpCreate, OpUpdate, OpDelete:
		return true
	}
	return false
}

type Op struct {
	Kind        OpKind
	TempID      string
	ID          string
	BaseVersion int64
	Data        map[ColumnID]any
}

type OpResult struct {
	Index       int
	Kind        OpKind
	TempID      string
	ID          string
	BaseVersion int64
}

type OpError struct {
	Index   int
	TempID  string
	ID      string
	Code    string
	Field   string
	Message string
}
