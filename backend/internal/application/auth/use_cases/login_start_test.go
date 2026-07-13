package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func newLoginStartDeps(t *testing.T, accs *accountsStub, creds *credentialsStub) usecases.LoginStartDeps {
	t.Helper()
	roles := newRolesStub()
	for id := range accs.byID {
		roles.put(id, iam.RoleExpert)
	}
	return usecases.LoginStartDeps{
		Accounts:    accs,
		Credentials: creds,
		Roles:       roles,
		Hasher:      &hasherStub{},
		Sessions:    &sessionsStub{createReply: "sess-tok"},
		SessionTTL:  time.Hour,
		Clock:       test.FrozenClock(),
	}
}

func TestLoginStart_HappyPath(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithEmail("user@example.com").Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	hash, err := vo.NewPasswordHash("hash")
	test.Must(t, err)
	cred, err := iam.NewCredentials(acc.ID(), hash)
	test.Must(t, err)
	test.Must(t, creds.Save(context.Background(), cred))
	creds.saveCalls = nil
	sess := &sessionsStub{createReply: "sess-tok"}
	deps := newLoginStartDeps(t, accs, creds)
	deps.Sessions = sess
	uc := usecases.NewLoginStartUseCase(deps)

	// act
	res, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "user@example.com", Password: "secret"})

	// assert
	test.NoErr(t, err)
	if res.Token != "sess-tok" {
		t.Fatalf("expected session token, got %q", res.Token)
	}
	if res.Account == nil || res.Account.ID() != acc.ID() {
		t.Fatalf("expected account in result")
	}
	if len(sess.createCalls) != 1 {
		t.Fatalf("expected one session create, got %d", len(sess.createCalls))
	}
}

func TestLoginStart_InvalidEmailReturnsUnauthorized(t *testing.T) {
	// arrange
	uc := usecases.NewLoginStartUseCase(newLoginStartDeps(t, newAccountsStub(), newCredentialsStub()))

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "not-an-email", Password: "secret"})

	// assert
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestLoginStart_AccountNotFoundReturnsUnauthorized(t *testing.T) {
	// arrange
	sess := &sessionsStub{}
	deps := newLoginStartDeps(t, newAccountsStub(), newCredentialsStub())
	deps.Sessions = sess
	uc := usecases.NewLoginStartUseCase(deps)

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "ghost@example.com", Password: "secret"})

	// assert
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if len(sess.createCalls) != 0 {
		t.Fatalf("expected no session when account is missing")
	}
}

func TestLoginStart_NotActiveReturnsUnauthorized(t *testing.T) {
	// arrange
	pending := iamtest.NewAccount(t).
		WithEmail("pending@example.com").
		WithStatus(iam.AccountStatusPendingPassword).
		Build(t)
	accs := newAccountsStub().put(pending)
	sess := &sessionsStub{}
	deps := newLoginStartDeps(t, accs, newCredentialsStub())
	deps.Sessions = sess
	uc := usecases.NewLoginStartUseCase(deps)

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "pending@example.com", Password: "secret"})

	// assert
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if len(sess.createCalls) != 0 {
		t.Fatalf("expected no session for non-active account")
	}
}

func TestLoginStart_WrongPasswordReturnsUnauthorized(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithEmail("user@example.com").Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	hash, _ := vo.NewPasswordHash("hash")
	cred, _ := iam.NewCredentials(acc.ID(), hash)
	test.Must(t, creds.Save(context.Background(), cred))
	sess := &sessionsStub{}
	deps := newLoginStartDeps(t, accs, creds)
	deps.Hasher = &hasherStub{verifyErr: errors.New("mismatch")}
	deps.Sessions = sess
	uc := usecases.NewLoginStartUseCase(deps)

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "user@example.com", Password: "wrong"})

	// assert
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if len(sess.createCalls) != 0 {
		t.Fatalf("expected no session on wrong password")
	}
}

func TestLoginStart_LockoutAfterFailures(t *testing.T) {
	acc := iamtest.NewAccount(t).WithEmail("user@example.com").Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	hash, _ := vo.NewPasswordHash("hash")
	cred, _ := iam.NewCredentials(acc.ID(), hash)
	test.Must(t, creds.Save(context.Background(), cred))
	attempts := newLoginAttemptsStub()
	deps := newLoginStartDeps(t, accs, creds)
	deps.Hasher = &hasherStub{verifyErr: errors.New("mismatch")}
	deps.Attempts = attempts
	deps.LockoutAttempts = 3
	deps.LockoutWindow = time.Minute
	uc := usecases.NewLoginStartUseCase(deps)

	for i := 0; i < 3; i++ {
		_, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "user@example.com", Password: "wrong"})
		if !errors.Is(err, shared.ErrUnauthorized) {
			t.Fatalf("expected unauthorized, got %v", err)
		}
	}
	if attempts.failCalls != 3 {
		t.Fatalf("expected 3 fail() calls, got %d", attempts.failCalls)
	}
	if !attempts.locked[acc.ID().String()] {
		t.Fatalf("expected account to be locked after threshold")
	}
}

func TestLoginStart_LockedAccountSkipsVerification(t *testing.T) {
	acc := iamtest.NewAccount(t).WithEmail("user@example.com").Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	hash, _ := vo.NewPasswordHash("hash")
	cred, _ := iam.NewCredentials(acc.ID(), hash)
	test.Must(t, creds.Save(context.Background(), cred))
	attempts := newLoginAttemptsStub()
	attempts.locked[acc.ID().String()] = true
	sess := &sessionsStub{}
	deps := newLoginStartDeps(t, accs, creds)
	deps.Attempts = attempts
	deps.LockoutAttempts = 3
	deps.LockoutWindow = time.Minute
	deps.Sessions = sess
	uc := usecases.NewLoginStartUseCase(deps)

	_, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "user@example.com", Password: "secret"})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if len(sess.createCalls) != 0 {
		t.Fatalf("no session must be issued when locked")
	}
	if attempts.failCalls != 0 {
		t.Fatalf("Fail must not be called when already locked")
	}
}

func TestLoginStart_SuccessResetsAttempts(t *testing.T) {
	acc := iamtest.NewAccount(t).WithEmail("user@example.com").Build(t)
	accs := newAccountsStub().put(acc)
	creds := newCredentialsStub()
	hash, _ := vo.NewPasswordHash("hash")
	cred, _ := iam.NewCredentials(acc.ID(), hash)
	test.Must(t, creds.Save(context.Background(), cred))
	attempts := newLoginAttemptsStub()
	deps := newLoginStartDeps(t, accs, creds)
	deps.Attempts = attempts
	deps.LockoutAttempts = 3
	deps.LockoutWindow = time.Minute
	uc := usecases.NewLoginStartUseCase(deps)

	_, err := uc.Execute(context.Background(), usecases.LoginStartCommand{Email: "user@example.com", Password: "secret"})
	test.NoErr(t, err)
	if attempts.resetCalls != 1 {
		t.Fatalf("expected Reset to be called once, got %d", attempts.resetCalls)
	}
}
