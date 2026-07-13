package apidate

import (
	"fmt"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type Date struct {
	t time.Time
}

func New(t time.Time) Date {
	if t.IsZero() {
		return Date{}
	}
	local := t.In(time.Local)
	return Date{t: time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)}
}

func Optional(t time.Time) *Date {
	if t.IsZero() {
		return nil
	}
	d := New(t)
	return &d
}

func (d Date) Time() time.Time { return d.t }

func (d Date) IsZero() bool { return d.t.IsZero() }

func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.t.Format(time.DateOnly)
}

func (Date) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        huma.TypeString,
		Description: "Calendar date. Accepts YYYY-MM-DD or RFC 3339 date-time (time part is truncated).",
		Examples:    []any{"2026-04-22"},
	}
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.t.Format(time.DateOnly) + `"`), nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		*d = Date{}
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	if s == "" {
		*d = Date{}
		return nil
	}
	t, err := parseFlexible(s)
	if err != nil {
		return err
	}
	*d = New(t)
	return nil
}

func (d Date) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return nil, nil
	}
	return []byte(d.t.Format(time.DateOnly)), nil
}

func (d *Date) UnmarshalText(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" {
		*d = Date{}
		return nil
	}
	t, err := parseFlexible(s)
	if err != nil {
		return err
	}
	*d = New(t)
	return nil
}

func parseFlexible(s string) (time.Time, error) {
	for _, layout := range []string{
		time.DateOnly,
		time.RFC3339,
		time.RFC3339Nano,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD or RFC 3339 date-time", s)
}
