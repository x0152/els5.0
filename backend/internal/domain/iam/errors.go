package iam

import (
	"fmt"

	"github.com/els/backend/internal/domain/shared"
)

var (
	ErrEmailTaken         = fmt.Errorf("email already taken: %w", shared.ErrConflict)
	ErrInvalidCredentials = fmt.Errorf("invalid credentials: %w", shared.ErrUnauthorized)
	ErrAccountBlocked     = fmt.Errorf("account is blocked: %w", shared.ErrForbidden)
	ErrAccountPending     = fmt.Errorf("account is pending password: %w", shared.ErrForbidden)
)
