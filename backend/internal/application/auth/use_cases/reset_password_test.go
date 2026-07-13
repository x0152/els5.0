package usecases_test

import (
	"context"
	"testing"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func newResetPasswordDeps(t *testing.T, accs *accountsStub, creds *credentialsStub, invites *invitesStub) usecases.ResetPasswordDeps {
	t.Helper()
	return usecases.ResetPasswordDeps{
		Accounts:    accs,
		Credentials: creds,
		Hasher:      &hasherStub{},
		Invites:     invites,
		Sessions:    &sessionsStub{},
		Policy:      iam.PasswordPolicy{MinLength: 8},
	}
}

func newResetPasswordDepsWithSessions(t *testing.T, accs *accountsStub, creds *credentialsStub, invites *invitesStub, sessions *sessionsStub) usecases.ResetPasswordDeps {
	t.Helper()
	d := newResetPasswordDeps(t, accs, creds, invites)
	d.Sessions = sessions
	return d
}

func TestResetPassword_HappyPathActiveAccountNoSession(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	invites := &invitesStub{
		consumeOK: ports.InviteToken{Purpose: ports.InviteTokenResetPassword, AccountID: acc.ID().String()},
	}
	uc := usecases.NewResetPasswordUseCase(newResetPasswordDeps(t, accs, creds, invites))

	// act
	res, err := uc.Execute(context.Background(), usecases.ResetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.NoErr(t, err)
	if res.Email != acc.Email().String() {
		t.Fatalf("expected email %q, got %q", acc.Email().String(), res.Email)
	}
	if len(creds.saveCalls) != 1 {
		t.Fatalf("expected credentials saved")
	}
	if len(accs.updateCalls) != 0 {
		t.Fatalf("expected no account status update for active account, got %d", len(accs.updateCalls))
	}
}

func TestResetPassword_PendingAccountActivated(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithStatus(iam.AccountStatusPendingPassword).Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	invites := &invitesStub{
		consumeOK: ports.InviteToken{Purpose: ports.InviteTokenResetPassword, AccountID: acc.ID().String()},
	}
	uc := usecases.NewResetPasswordUseCase(newResetPasswordDeps(t, accs, creds, invites))

	// act
	_, err := uc.Execute(context.Background(), usecases.ResetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.NoErr(t, err)
	if !acc.IsActive() {
		t.Fatalf("expected pending account activated")
	}
	if len(accs.updateCalls) != 1 {
		t.Fatalf("expected account update once")
	}
}

func TestResetPassword_PolicyFails(t *testing.T) {
	// arrange
	uc := usecases.NewResetPasswordUseCase(newResetPasswordDeps(t, newAccountsStub(), newCredentialsStub(), &invitesStub{}))

	// act
	_, err := uc.Execute(context.Background(), usecases.ResetPasswordCommand{
		Token: "tok", Password: "x", PasswordConfirm: "x",
	})

	// assert
	test.ErrIs(t, err, shared.ErrValidation)
}

func TestResetPassword_WrongPurposeUnauthorized(t *testing.T) {
	// arrange
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenSetPassword, AccountID: "x"}}
	uc := usecases.NewResetPasswordUseCase(newResetPasswordDeps(t, newAccountsStub(), newCredentialsStub(), invites))

	// act
	_, err := uc.Execute(context.Background(), usecases.ResetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.ErrIs(t, err, shared.ErrUnauthorized)
}

func TestResetPassword_RevokesSessionsOnSuccess(t *testing.T) {
	acc := iamtest.NewAccount(t).Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	invites := &invitesStub{
		consumeOK: ports.InviteToken{Purpose: ports.InviteTokenResetPassword, AccountID: acc.ID().String()},
	}
	sessions := &sessionsStub{}
	uc := usecases.NewResetPasswordUseCase(newResetPasswordDepsWithSessions(t, accs, creds, invites, sessions))

	_, err := uc.Execute(context.Background(), usecases.ResetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	test.NoErr(t, err)
	if len(sessions.revokedAccounts) != 1 || sessions.revokedAccounts[0] != acc.ID().String() {
		t.Fatalf("expected RevokeByAccountID(%s), got %v", acc.ID().String(), sessions.revokedAccounts)
	}
}

func TestResetPassword_BlockedForbidden(t *testing.T) {
	// arrange
	blocked := iamtest.NewAccount(t).WithStatus(iam.AccountStatusBlocked).Build(t)
	accs := newAccountsStub().put(blocked)
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenResetPassword, AccountID: blocked.ID().String()}}
	uc := usecases.NewResetPasswordUseCase(newResetPasswordDeps(t, accs, newCredentialsStub(), invites))

	// act
	_, err := uc.Execute(context.Background(), usecases.ResetPasswordCommand{
		Token: "tok", Password: "password1", PasswordConfirm: "password1",
	})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
}
