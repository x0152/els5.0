package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func newResendUC(accs *accountsStub, invites *invitesStub, mail *mailStub) *usecases.ResendInviteUseCase {
	return usecases.NewResendInviteUseCase(usecases.ResendInviteDeps{
		Accounts: accs,
		Invites:  invites,
		Mail:     mail,
		TTL:      time.Hour,
		LinkTmpl: "https://app/{token}",
	})
}

func TestResendInvite_HappyPath(t *testing.T) {
	// arrange
	pending := iamtest.NewAccount(t).WithEmail("pending@example.com").WithStatus(iam.AccountStatusPendingPassword).Build(t)
	accs := newAccountsStub().put(pending)
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := newResendUC(accs, invites, mail)

	// act
	res, err := uc.Execute(context.Background(), usecases.ResendInviteCommand{Email: "pending@example.com"})

	// assert
	test.NoErr(t, err)
	if !strings.Contains(res.SentTo, "***") {
		t.Fatalf("expected masked email, got %q", res.SentTo)
	}
	if len(invites.issued) != 1 || len(mail.invites) != 1 {
		t.Fatalf("expected invite issued and mailed")
	}
	if mail.invites[0].Name != pending.Name().Full() {
		t.Fatalf("expected full recipient name %q, got %q", pending.Name().Full(), mail.invites[0].Name)
	}
}

func TestResendInvite_InvalidEmailMaskQuietly(t *testing.T) {
	// arrange
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := newResendUC(newAccountsStub(), invites, mail)

	// act
	_, err := uc.Execute(context.Background(), usecases.ResendInviteCommand{Email: "broken"})

	// assert
	test.NoErr(t, err)
	if len(invites.issued) != 0 || len(mail.invites) != 0 {
		t.Fatalf("expected no side effects")
	}
}

func TestResendInvite_NotFoundMaskQuietly(t *testing.T) {
	// arrange
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := newResendUC(newAccountsStub(), invites, mail)

	// act
	_, err := uc.Execute(context.Background(), usecases.ResendInviteCommand{Email: "ghost@example.com"})

	// assert
	test.NoErr(t, err)
	if len(invites.issued) != 0 || len(mail.invites) != 0 {
		t.Fatalf("expected no side effects when account is missing")
	}
}

func TestResendInvite_ActiveAccountMaskQuietly(t *testing.T) {
	// arrange
	active := iamtest.NewAccount(t).WithEmail("active@example.com").Build(t)
	accs := newAccountsStub().put(active)
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := newResendUC(accs, invites, mail)

	// act
	_, err := uc.Execute(context.Background(), usecases.ResendInviteCommand{Email: "active@example.com"})

	// assert
	test.NoErr(t, err)
	if len(invites.issued) != 0 || len(mail.invites) != 0 {
		t.Fatalf("expected no side effects when account is already active")
	}
}

func TestResendInvite_GetByEmailFailurePropagates(t *testing.T) {
	// arrange
	accs := newAccountsStub()
	accs.getByEmailErr = errors.New("db down")
	uc := newResendUC(accs, &invitesStub{}, &mailStub{})

	// act
	_, err := uc.Execute(context.Background(), usecases.ResendInviteCommand{Email: "user@example.com"})

	// assert
	if err == nil {
		t.Fatalf("expected error to propagate")
	}
}

func TestResendInvite_MailFailurePropagates(t *testing.T) {
	// arrange
	pending := iamtest.NewAccount(t).WithEmail("pending@example.com").WithStatus(iam.AccountStatusPendingPassword).Build(t)
	accs := newAccountsStub().put(pending)
	mail := &mailStub{inviteErr: errors.New("smtp")}
	uc := newResendUC(accs, &invitesStub{}, mail)

	// act
	_, err := uc.Execute(context.Background(), usecases.ResendInviteCommand{Email: "pending@example.com"})

	// assert
	if err == nil {
		t.Fatalf("expected mail error to propagate")
	}
}
