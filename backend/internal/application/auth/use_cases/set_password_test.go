package usecases_test

import (
	"context"
	"errors"
	"testing"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func newSetPasswordDeps(t *testing.T, accs *accountsStub, creds *credentialsStub, invites *invitesStub) usecases.SetPasswordDeps {
	t.Helper()
	return usecases.SetPasswordDeps{
		Accounts:    accs,
		Credentials: creds,
		Hasher:      &hasherStub{},
		Invites:     invites,
		Sessions:    &sessionsStub{},
		Policy:      iam.PasswordPolicy{MinLength: 8},
	}
}

func newSetPasswordDepsWithSessions(t *testing.T, accs *accountsStub, creds *credentialsStub, invites *invitesStub, sessions *sessionsStub) usecases.SetPasswordDeps {
	t.Helper()
	d := newSetPasswordDeps(t, accs, creds, invites)
	d.Sessions = sessions
	return d
}

func TestSetPassword_HappyPathActivatesWithoutSession(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithStatus(iam.AccountStatusPendingPassword).Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	invites := &invitesStub{
		consumeOK: ports.InviteToken{Purpose: ports.InviteTokenSetPassword, AccountID: acc.ID().String()},
	}
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDeps(t, accs, creds, invites))

	// act
	res, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.NoErr(t, err)
	if res.Email != acc.Email().String() {
		t.Fatalf("expected email %q, got %q", acc.Email().String(), res.Email)
	}
	if !acc.IsActive() {
		t.Fatalf("expected account to be activated")
	}
	if len(creds.saveCalls) != 1 {
		t.Fatalf("expected credentials saved once")
	}
	if len(accs.updateCalls) != 1 {
		t.Fatalf("expected account update")
	}
}

func TestSetPassword_PolicyFails(t *testing.T) {
	// arrange
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDeps(t, newAccountsStub(), newCredentialsStub(), &invitesStub{}))

	// act
	_, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "short", PasswordConfirm: "short",
	})

	// assert
	test.ErrIs(t, err, shared.ErrValidation)
}

func TestSetPassword_TokenConsumeFails(t *testing.T) {
	// arrange
	invites := &invitesStub{consumeErr: errors.New("expired")}
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDeps(t, newAccountsStub(), newCredentialsStub(), invites))

	// act
	_, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	if err == nil {
		t.Fatalf("expected error from consume")
	}
}

func TestSetPassword_WrongPurposeUnauthorized(t *testing.T) {
	// arrange
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenMagicLogin, AccountID: "x"}}
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDeps(t, newAccountsStub(), newCredentialsStub(), invites))

	// act
	_, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.ErrIs(t, err, shared.ErrUnauthorized)
}

func TestSetPassword_ActiveAccountOverwritesIdempotently(t *testing.T) {
	// arrange: already-active account (e.g. activated via reset-password),
	// but the invite token is still valid — password setup must succeed, not fail with 409.
	active := iamtest.NewAccount(t).Build(t)
	accs := newAccountsStub().put(active)
	creds := newCredentialsStub()
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenSetPassword, AccountID: active.ID().String()}}
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDeps(t, accs, creds, invites))

	// act
	res, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.NoErr(t, err)
	if res.Email != active.Email().String() {
		t.Fatalf("expected email %q, got %q", active.Email().String(), res.Email)
	}
	if len(creds.saveCalls) != 1 {
		t.Fatalf("expected credentials saved once, got %d", len(creds.saveCalls))
	}
	if len(accs.updateCalls) != 0 {
		t.Fatalf("expected no account update for already-active account, got %d", len(accs.updateCalls))
	}
}

func TestSetPassword_BlockedAccountForbidden(t *testing.T) {
	// arrange
	blocked := iamtest.NewAccount(t).WithStatus(iam.AccountStatusBlocked).Build(t)
	accs := newAccountsStub().put(blocked)
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenSetPassword, AccountID: blocked.ID().String()}}
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDeps(t, accs, newCredentialsStub(), invites))

	// act
	_, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestSetPassword_RevokesSessionsOnSuccess(t *testing.T) {
	acc := iamtest.NewAccount(t).WithStatus(iam.AccountStatusPendingPassword).Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	invites := &invitesStub{
		consumeOK: ports.InviteToken{Purpose: ports.InviteTokenSetPassword, AccountID: acc.ID().String()},
	}
	sessions := &sessionsStub{}
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDepsWithSessions(t, accs, creds, invites, sessions))

	_, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	test.NoErr(t, err)
	if len(sessions.revokedAccounts) != 1 || sessions.revokedAccounts[0] != acc.ID().String() {
		t.Fatalf("expected RevokeByAccountID(%s), got %v", acc.ID().String(), sessions.revokedAccounts)
	}
}

func TestSetPassword_CredentialsSaveFails(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithStatus(iam.AccountStatusPendingPassword).Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	creds.saveErr = errors.New("db down")
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenSetPassword, AccountID: acc.ID().String()}}
	uc := usecases.NewSetPasswordUseCase(newSetPasswordDeps(t, accs, creds, invites))

	// act
	_, err := uc.Execute(context.Background(), usecases.SetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	if err == nil || err.Error() != "db down" {
		t.Fatalf("expected creds save error, got %v", err)
	}
}
