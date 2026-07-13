package admin

import (
	"fmt"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

type ID struct{ vo.ID }

func NewID() ID { return ID{ID: vo.NewID()} }

type Administrator struct {
	iam.AccountSide
	id         ID
	timestamps vo.Timestamps
}

type NewAdministratorParams struct {
	ID         ID
	Account    *iam.Account
	Timestamps vo.Timestamps
}

func NewAdministrator(p NewAdministratorParams) (*Administrator, error) {
	side, err := iam.NewAccountSide(p.Account)
	if err != nil {
		return nil, err
	}
	a := &Administrator{
		AccountSide: side,
		id:          p.ID,
		timestamps:  p.Timestamps,
	}
	if err := a.validate(); err != nil {
		return nil, err
	}
	return a, nil
}

func NewAdministratorNow(id ID, account *iam.Account) (*Administrator, error) {
	timestamps, err := vo.NewCurrentTimestamps()
	if err != nil {
		return nil, shared.Validation(fmt.Errorf("administrator.timestamps: %w", err))
	}
	return NewAdministrator(NewAdministratorParams{
		ID:         id,
		Account:    account,
		Timestamps: timestamps,
	})
}

func (a *Administrator) ID() ID               { return a.id }
func (a *Administrator) CreatedAt() time.Time { return a.timestamps.CreatedAt() }
func (a *Administrator) UpdatedAt() time.Time { return a.timestamps.UpdatedAt() }

func (a *Administrator) Version() int64 {
	roleV := a.timestamps.UpdatedAt().UnixNano()
	accV := a.Account().UpdatedAt().UnixNano()
	if accV > roleV {
		return accV
	}
	return roleV
}

func (a *Administrator) validate() error {
	var errs []error
	if a.id.IsZero() {
		errs = append(errs, fmt.Errorf("administrator.id: must not be zero"))
	}
	if a.Account() == nil || a.AccountID().IsZero() {
		errs = append(errs, fmt.Errorf("administrator.account: must not be zero"))
	}
	if _, err := vo.NewTimestamps(a.timestamps.CreatedAt(), a.timestamps.UpdatedAt()); err != nil {
		errs = append(errs, fmt.Errorf("administrator.timestamps: %w", err))
	}
	return shared.Validation(errs...)
}
