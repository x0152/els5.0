package vo

import (
	"fmt"

	"github.com/google/uuid"
)

type ID struct {
	v uuid.UUID
}

func NewID() ID                 { return ID{v: uuid.New()} }
func IDFromUUID(u uuid.UUID) ID { return ID{v: u} }

func ParseID(s string) (ID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("invalid id: %w", err)
	}
	return ID{v: u}, nil
}

func (i ID) String() string  { return i.v.String() }
func (i ID) IsZero() bool    { return i.v == uuid.Nil }
func (i ID) UUID() uuid.UUID { return i.v }
