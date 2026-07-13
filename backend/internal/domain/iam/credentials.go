package iam

import (
	"fmt"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/domain/shared/vo"
)

type Credentials struct {
	accountID AccountID
	hash      vo.PasswordHash
}

func NewCredentials(accountID AccountID, hash vo.PasswordHash) (*Credentials, error) {
	var errs []error
	if accountID.IsZero() {
		errs = append(errs, fmt.Errorf("credentials.account_id: must not be zero"))
	}
	if hash.IsZero() {
		errs = append(errs, fmt.Errorf("credentials.hash: must not be empty"))
	}
	if err := shared.Validation(errs...); err != nil {
		return nil, err
	}
	return &Credentials{accountID: accountID, hash: hash}, nil
}

func HydrateCredentials(accountID AccountID, hash vo.PasswordHash) *Credentials {
	return &Credentials{accountID: accountID, hash: hash}
}

func (c *Credentials) AccountID() AccountID  { return c.accountID }
func (c *Credentials) Hash() vo.PasswordHash { return c.hash }

func (c *Credentials) Verify(plain string, h ports.PasswordHasher) error {
	if err := h.Verify(c.hash, plain); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}
