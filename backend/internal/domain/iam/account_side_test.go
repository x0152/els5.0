package iam_test

import (
	"testing"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
)

func TestNewAccountSide(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)

	// act
	side, err := iam.NewAccountSide(acc)

	// assert
	test.NoErr(t, err)
	if side.Account() != acc || side.AccountID() != acc.ID() {
		t.Fatalf("side must reference original account")
	}
}

func TestNewAccountSide_RejectsNil(t *testing.T) {
	_, err := iam.NewAccountSide(nil)
	test.ErrIs(t, err, shared.ErrValidation)
}

func TestAccountSide_ChangeEmail(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)

	// act + assert
	test.NoErr(t, side.ChangeEmail("new@example.com"))
	if acc.Email().String() != "new@example.com" {
		t.Fatalf("expected email updated, got %s", acc.Email())
	}
	test.ErrIs(t, side.ChangeEmail("broken"), shared.ErrValidation)
}

func TestAccountSide_Rename(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)

	// act + assert
	test.NoErr(t, side.Rename("Jane", "Smith"))
	if acc.Name().Full() != "Jane Smith" {
		t.Fatalf("expected name updated, got %s", acc.Name().Full())
	}
	test.ErrIs(t, side.Rename("", ""), shared.ErrValidation)
}

func TestAccountSide_SetStatus(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)

	// act + assert: active -> blocked
	test.NoErr(t, side.SetStatus(iam.AccountStatusBlocked))
	if acc.Status() != iam.AccountStatusBlocked {
		t.Fatalf("expected blocked, got %s", acc.Status())
	}

	// blocked -> active goes through Unblock
	test.NoErr(t, side.SetStatus(iam.AccountStatusActive))
	if acc.Status() != iam.AccountStatusActive {
		t.Fatalf("expected active, got %s", acc.Status())
	}

	// idempotent same-state
	test.NoErr(t, side.SetStatus(iam.AccountStatusActive))

	// invalid status
	test.ErrIs(t, side.SetStatus("garbage"), shared.ErrValidation)
}

func TestAccountSide_SetStatus_NoAuthToPendingPassword(t *testing.T) {
	// arrange
	params := validParams()
	params.Status = iam.AccountStatusNoAuth
	acc, err := iam.NewAccount(params)
	test.NoErr(t, err)
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)

	// act
	err = side.SetStatus(iam.AccountStatusPendingPassword)

	// assert
	test.NoErr(t, err)
	if acc.Status() != iam.AccountStatusPendingPassword {
		t.Fatalf("expected pending_password, got %s", acc.Status())
	}
	if err := acc.EnsureCanLogin(); err == nil {
		t.Fatal("no_auth/pending account must not be able to login")
	}
}

func TestAccountSide_SetStatus_RejectsManualNoAuthTransition(t *testing.T) {
	for _, status := range []iam.AccountStatus{
		iam.AccountStatusActive,
		iam.AccountStatusBlocked,
		iam.AccountStatusPendingPassword,
	} {
		t.Run(status.String(), func(t *testing.T) {
			// arrange
			params := validParams()
			params.Status = status
			acc, err := iam.NewAccount(params)
			test.NoErr(t, err)
			side, err := iam.NewAccountSide(acc)
			test.NoErr(t, err)

			// act
			err = side.SetStatus(iam.AccountStatusNoAuth)

			// assert
			test.ErrIs(t, err, shared.ErrConflict)
		})
	}
}
