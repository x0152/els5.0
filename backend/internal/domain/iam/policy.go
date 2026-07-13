package iam

import (
	"fmt"

	"github.com/els/backend/internal/domain/shared"
)

const (
	defaultPasswordMinLength = 8
	defaultPasswordMaxLength = 128
	hardPasswordMaxLength    = 1024
)

type PasswordPolicy struct {
	MinLength int
	MaxLength int
}

func (p PasswordPolicy) Validate(plain string) error {
	minLen := p.MinLength
	if minLen <= 0 {
		minLen = defaultPasswordMinLength
	}
	maxLen := p.MaxLength
	if maxLen <= 0 {
		maxLen = defaultPasswordMaxLength
	}
	if maxLen > hardPasswordMaxLength {
		maxLen = hardPasswordMaxLength
	}
	if len(plain) < minLen {
		return shared.Validation(fmt.Errorf("password: must be at least %d characters", minLen))
	}
	if len(plain) > maxLen {
		return shared.Validation(fmt.Errorf("password: must be at most %d bytes", maxLen))
	}
	return nil
}

func (p PasswordPolicy) Compare(password, confirm string) error {
	if err := p.Validate(password); err != nil {
		return err
	}
	if password != confirm {
		return shared.Validation(fmt.Errorf("password_confirm: must match password"))
	}
	return nil
}
