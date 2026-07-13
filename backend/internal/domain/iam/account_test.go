package iam_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
)

func validParams() iam.NewAccountParams {
	ts, _ := vo.NewTimestamps(test.FixedTime, test.FixedTime)
	return iam.NewAccountParams{
		ID:         iam.NewAccountID(),
		Email:      "user@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		Status:     iam.AccountStatusActive,
		Timestamps: ts,
	}
}

func TestNewAccount(t *testing.T) {
	cases := []struct {
		name    string
		mut     func(*iam.NewAccountParams)
		wantErr error
	}{
		{name: "ok", mut: func(p *iam.NewAccountParams) {}},
		{name: "zero_id", mut: func(p *iam.NewAccountParams) { p.ID = iam.AccountID{} }, wantErr: shared.ErrValidation},
		{name: "bad_email", mut: func(p *iam.NewAccountParams) { p.Email = "not-an-email" }, wantErr: shared.ErrValidation},
		{name: "empty_name", mut: func(p *iam.NewAccountParams) { p.FirstName = ""; p.LastName = "" }, wantErr: shared.ErrValidation},
		{name: "invalid_status", mut: func(p *iam.NewAccountParams) { p.Status = "garbage" }, wantErr: shared.ErrValidation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := validParams()
			tc.mut(&p)

			acc, err := iam.NewAccount(p)

			if tc.wantErr != nil {
				test.ErrIs(t, err, tc.wantErr)
				return
			}
			test.NoErr(t, err)
			if acc.ID() != p.ID || acc.Status() != p.Status {
				t.Errorf("unexpected account id/status: %v/%s", acc.ID(), acc.Status())
			}
		})
	}
}

func TestNewPendingAccountNow(t *testing.T) {
	id := iam.NewAccountID()
	acc, err := iam.NewPendingAccountNow(id, "u@example.com", "John", "Doe")

	test.NoErr(t, err)
	if acc.Status() != iam.AccountStatusPendingPassword {
		t.Errorf("expected pending status, got %s", acc.Status())
	}
}

func TestAccount_LifecycleTransitions(t *testing.T) {
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)

	if err := acc.Block(); err != nil {
		t.Fatalf("block: %v", err)
	}
	if !errors.Is(acc.EnsureCanLogin(), iam.ErrAccountBlocked) {
		t.Errorf("blocked must not be able to login")
	}
	if err := acc.Block(); err != nil {
		t.Errorf("block on blocked must be no-op, got %v", err)
	}
	if err := acc.Unblock(); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	if acc.Status() != iam.AccountStatusActive {
		t.Errorf("expected active after unblock, got %s", acc.Status())
	}
}

func TestAccount_Activate(t *testing.T) {
	acc, err := iam.NewPendingAccountNow(iam.NewAccountID(), "u@example.com", "John", "Doe")
	test.NoErr(t, err)
	if err := acc.EnsureCanLogin(); !errors.Is(err, iam.ErrAccountPending) {
		t.Errorf("pending account must not login, got %v", err)
	}
	if err := acc.Activate(); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !acc.IsActive() {
		t.Errorf("expected active after Activate")
	}
}

func TestAccount_Activate_FromBlockedConflict(t *testing.T) {
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	test.NoErr(t, acc.Block())

	err = acc.Activate()

	test.ErrIs(t, err, shared.ErrConflict)
}

func TestAccount_Rename(t *testing.T) {
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	originalUpdated := acc.UpdatedAt()
	time.Sleep(time.Millisecond)

	newName, err := vo.NewPersonName("Jane", "Smith")
	test.NoErr(t, err)
	test.NoErr(t, acc.Rename(newName))

	if acc.Name().Full() != "Jane Smith" {
		t.Errorf("expected new name, got %s", acc.Name().Full())
	}
	if !acc.UpdatedAt().After(originalUpdated) {
		t.Errorf("expected updated_at to advance")
	}
}

func TestAccount_Rename_RejectsZero(t *testing.T) {
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)

	err = acc.Rename(vo.PersonName{})

	test.ErrIs(t, err, shared.ErrValidation)
}

func TestAccount_ChangeEmail(t *testing.T) {
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)

	t.Run("noop_when_same", func(t *testing.T) {
		updatedBefore := acc.UpdatedAt()
		email, err := vo.NewEmail("user@example.com")
		test.NoErr(t, err)
		test.NoErr(t, acc.ChangeEmail(email))
		if !acc.UpdatedAt().Equal(updatedBefore) {
			t.Errorf("expected no-op for same email")
		}
	})

	t.Run("rejects_zero_email", func(t *testing.T) {
		err := acc.ChangeEmail(vo.Email{})
		test.ErrIs(t, err, shared.ErrValidation)
	})

	t.Run("ok_changes", func(t *testing.T) {
		newEmail, err := vo.NewEmail("new@example.com")
		test.NoErr(t, err)
		test.NoErr(t, acc.ChangeEmail(newEmail))
		if acc.Email().String() != "new@example.com" {
			t.Errorf("expected new email, got %s", acc.Email())
		}
	})
}

func TestAccount_ChangePictureURL(t *testing.T) {
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)

	test.NoErr(t, acc.ChangePictureURL("  https://cdn/pic.png  "))

	if acc.PictureURL() != "https://cdn/pic.png" {
		t.Errorf("expected trimmed URL, got %q", acc.PictureURL())
	}
}

func TestParseAccountStatus(t *testing.T) {
	cases := []struct {
		in   string
		want iam.AccountStatus
		ok   bool
	}{
		{in: "active", want: iam.AccountStatusActive, ok: true},
		{in: "blocked", want: iam.AccountStatusBlocked, ok: true},
		{in: "pending_password", want: iam.AccountStatusPendingPassword, ok: true},
		{in: "no_auth", want: iam.AccountStatusNoAuth, ok: true},
		{in: "garbage", ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := iam.ParseAccountStatus(tc.in)
			if tc.ok {
				test.NoErr(t, err)
				if got != tc.want {
					t.Errorf("expected %s, got %s", tc.want, got)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "invalid") {
				t.Errorf("expected invalid error, got %v", err)
			}
		})
	}
}
