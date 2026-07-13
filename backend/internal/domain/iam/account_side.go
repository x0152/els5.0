package iam

import (
	"fmt"
	"time"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

type AccountSide struct {
	account *Account
}

func NewAccountSide(a *Account) (AccountSide, error) {
	if a == nil {
		return AccountSide{}, shared.Validation(fmt.Errorf("account: must not be nil"))
	}
	return AccountSide{account: a}, nil
}

func (s AccountSide) Account() *Account           { return s.account }
func (s AccountSide) AccountID() AccountID        { return s.account.ID() }
func (s AccountSide) Email() vo.Email             { return s.account.Email() }
func (s AccountSide) Name() vo.PersonName         { return s.account.Name() }
func (s AccountSide) FirstName() string           { return s.account.Name().First() }
func (s AccountSide) LastName() string            { return s.account.Name().Last() }
func (s AccountSide) Status() AccountStatus       { return s.account.Status() }
func (s AccountSide) AccountCreatedAt() time.Time { return s.account.CreatedAt() }
func (s AccountSide) AccountUpdatedAt() time.Time { return s.account.UpdatedAt() }

func (s AccountSide) ChangeEmail(raw string) error {
	email, err := vo.NewEmail(raw)
	if err != nil {
		return shared.Validation(fmt.Errorf("account.email: %w", err))
	}
	return s.account.ChangeEmail(email)
}

func (s AccountSide) Rename(first, last string) error {
	name, err := vo.NewPersonName(first, last)
	if err != nil {
		return shared.Validation(fmt.Errorf("account.name: %w", err))
	}
	return s.account.Rename(name)
}

func (s AccountSide) SetStatus(status AccountStatus) error {
	if !status.IsValid() {
		return shared.Validation(fmt.Errorf("account.status: invalid %q", status))
	}
	if s.account.Status() == status {
		return nil
	}
	switch status {
	case AccountStatusActive:
		if s.account.Status() == AccountStatusBlocked {
			return s.account.Unblock()
		}
		return s.account.Activate()
	case AccountStatusPendingPassword:
		if s.account.Status() == AccountStatusNoAuth {
			return s.account.transitionTo(AccountStatusPendingPassword)
		}
		return fmt.Errorf("%w: cannot set account from %s to %s", shared.ErrConflict, s.account.Status(), status)
	case AccountStatusBlocked:
		return s.account.Block()
	case AccountStatusNoAuth:
		return fmt.Errorf("%w: cannot set account to %s", shared.ErrConflict, status)
	}
	return fmt.Errorf("%w: cannot set account to %s", shared.ErrConflict, status)
}
