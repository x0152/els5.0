package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestForgotPassword_HappyPath(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithEmail("user@example.com").Build(t)
	accs := newAccountsStub().put(acc)
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := usecases.NewForgotPasswordUseCase(usecases.ForgotPasswordDeps{
		Accounts: accs,
		Invites:  invites,
		Mail:     mail,
		TTL:      time.Hour,
		LinkTmpl: "https://app/{token}",
	})

	// act
	res, err := uc.Execute(context.Background(), usecases.ForgotPasswordCommand{Email: "user@example.com"})

	// assert
	test.NoErr(t, err)
	if !strings.Contains(res.SentTo, "***") {
		t.Fatalf("expected masked email, got %q", res.SentTo)
	}
	if len(invites.issued) != 1 || invites.issued[0].Purpose != ports.InviteTokenResetPassword {
		t.Fatalf("expected reset invite issued, got %+v", invites.issued)
	}
	if len(mail.resets) != 1 || !strings.Contains(mail.resets[0].Link, "issued-token-reset_password") {
		t.Fatalf("expected reset mail with token, got %+v", mail.resets)
	}
}

func TestForgotPassword_InvalidEmailMaskQuietly(t *testing.T) {
	// arrange
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := usecases.NewForgotPasswordUseCase(usecases.ForgotPasswordDeps{
		Accounts: newAccountsStub(),
		Invites:  invites,
		Mail:     mail,
	})

	// act
	res, err := uc.Execute(context.Background(), usecases.ForgotPasswordCommand{Email: "broken"})

	// assert
	test.NoErr(t, err)
	if res.SentTo != "***" {
		t.Fatalf("expected mask, got %q", res.SentTo)
	}
	if len(invites.issued) != 0 || len(mail.resets) != 0 {
		t.Fatalf("expected no side effects on invalid email")
	}
}

func TestForgotPassword_NotFoundMaskQuietly(t *testing.T) {
	// arrange
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := usecases.NewForgotPasswordUseCase(usecases.ForgotPasswordDeps{
		Accounts: newAccountsStub(),
		Invites:  invites,
		Mail:     mail,
	})

	// act
	_, err := uc.Execute(context.Background(), usecases.ForgotPasswordCommand{Email: "ghost@example.com"})

	// assert
	test.NoErr(t, err)
	if len(invites.issued) != 0 || len(mail.resets) != 0 {
		t.Fatalf("expected no side effects when account is missing")
	}
}

func TestForgotPassword_BlockedMaskQuietly(t *testing.T) {
	// arrange
	blocked := iamtest.NewAccount(t).WithEmail("blocked@example.com").WithStatus(iam.AccountStatusBlocked).Build(t)
	accs := newAccountsStub().put(blocked)
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := usecases.NewForgotPasswordUseCase(usecases.ForgotPasswordDeps{
		Accounts: accs,
		Invites:  invites,
		Mail:     mail,
	})

	// act
	_, err := uc.Execute(context.Background(), usecases.ForgotPasswordCommand{Email: "blocked@example.com"})

	// assert
	test.NoErr(t, err)
	if len(invites.issued) != 0 || len(mail.resets) != 0 {
		t.Fatalf("expected no side effects on blocked account")
	}
}

func TestForgotPassword_NoAuthMaskQuietly(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithEmail("seed@example.com").WithStatus(iam.AccountStatusNoAuth).Build(t)
	accs := newAccountsStub().put(acc)
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := usecases.NewForgotPasswordUseCase(usecases.ForgotPasswordDeps{
		Accounts: accs,
		Invites:  invites,
		Mail:     mail,
	})

	// act
	_, err := uc.Execute(context.Background(), usecases.ForgotPasswordCommand{Email: "seed@example.com"})

	// assert
	test.NoErr(t, err)
	if len(invites.issued) != 0 || len(mail.resets) != 0 {
		t.Fatalf("expected no side effects on no_auth account")
	}
}

func TestForgotPassword_MailFailurePropagates(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).WithEmail("user@example.com").Build(t)
	accs := newAccountsStub().put(acc)
	mail := &mailStub{resetErr: errors.New("smtp")}
	uc := usecases.NewForgotPasswordUseCase(usecases.ForgotPasswordDeps{
		Accounts: accs,
		Invites:  &invitesStub{},
		Mail:     mail,
	})

	// act
	_, err := uc.Execute(context.Background(), usecases.ForgotPasswordCommand{Email: "user@example.com"})

	// assert
	if err == nil {
		t.Fatalf("expected mail error to propagate")
	}
}
