package vo

import (
	"errors"
	"fmt"
	"strings"
)

type PersonName struct {
	first string
	last  string
}

func NewPersonName(first, last string) (PersonName, error) {
	f := strings.TrimSpace(first)
	l := strings.TrimSpace(last)

	var errs []error
	if f == "" {
		errs = append(errs, fmt.Errorf("first_name must not be empty"))
	}
	if l == "" {
		errs = append(errs, fmt.Errorf("last_name must not be empty"))
	}
	if err := errors.Join(errs...); err != nil {
		return PersonName{}, err
	}
	return PersonName{first: f, last: l}, nil
}

func (p PersonName) First() string { return p.first }
func (p PersonName) Last() string  { return p.last }
func (p PersonName) Full() string  { return p.first + " " + p.last }
func (p PersonName) IsZero() bool  { return p.first == "" && p.last == "" }
