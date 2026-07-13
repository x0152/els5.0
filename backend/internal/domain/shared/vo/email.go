package vo

import (
	"fmt"
	"net/mail"
	"strings"
)

type Email struct {
	v string
}

func NewEmail(raw string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return Email{}, fmt.Errorf("email must not be empty")
	}
	addr, err := mail.ParseAddress(normalized)
	if err != nil || addr.Address != normalized {
		return Email{}, fmt.Errorf("invalid email: %q", raw)
	}
	return Email{v: normalized}, nil
}

func (e Email) String() string { return e.v }
func (e Email) IsZero() bool   { return e.v == "" }
