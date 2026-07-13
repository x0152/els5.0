package vo

import "fmt"

type PasswordHash struct {
	v string
}

func NewPasswordHash(raw string) (PasswordHash, error) {
	if raw == "" {
		return PasswordHash{}, fmt.Errorf("password_hash must not be empty")
	}
	return PasswordHash{v: raw}, nil
}

func (p PasswordHash) String() string { return p.v }
func (p PasswordHash) IsZero() bool   { return p.v == "" }
