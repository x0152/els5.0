package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func newInviteUC(accs *accountsStub, invites *invitesStub, mail *mailStub) *usecases.InviteAccountUseCase {
	return usecases.NewInviteAccountUseCase(usecases.InviteAccountDeps{
		Accounts: accs,
		Invites:  invites,
		Mail:     mail,
		TTL:      time.Hour,
		LinkTmpl: "https://app/{token}",
	})
}

func TestInviteAccount_HappyPath(t *testing.T) {
	// arrange
	accs := newAccountsStub()
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := newInviteUC(accs, invites, mail)

	// act
	acc, err := uc.Execute(context.Background(), usecases.InviteAccountCommand{
		Email: "new@example.com", FirstName: "Jane", LastName: "Roe",
	})

	// assert
	test.NoErr(t, err)
	if acc == nil || acc.Email().String() != "new@example.com" {
		t.Fatalf("expected account created with email, got %+v", acc)
	}
	if acc.Status() != iam.AccountStatusPendingPassword {
		t.Fatalf("expected pending_password, got %s", acc.Status())
	}
	if len(invites.issued) != 1 || invites.issued[0].Purpose != ports.InviteTokenSetPassword {
		t.Fatalf("expected set_password invite, got %+v", invites.issued)
	}
	if len(mail.invites) != 1 || !strings.Contains(mail.invites[0].Link, "issued-token-set_password") {
		t.Fatalf("expected invite mail with token, got %+v", mail.invites)
	}
	if mail.invites[0].Name != "Jane Roe" {
		t.Fatalf("expected full recipient name Jane Roe, got %q", mail.invites[0].Name)
	}
}

func TestInviteAccount_ValidationError(t *testing.T) {
	// arrange
	uc := newInviteUC(newAccountsStub(), &invitesStub{}, &mailStub{})

	// act
	_, err := uc.Execute(context.Background(), usecases.InviteAccountCommand{Email: "broken"})

	// assert
	test.ErrIs(t, err, shared.ErrValidation)
}

func TestInviteAccount_RepoCreateFails(t *testing.T) {
	// arrange
	accs := newAccountsStub()
	accs.createErr = errors.New("conflict")
	uc := newInviteUC(accs, &invitesStub{}, &mailStub{})

	// act
	_, err := uc.Execute(context.Background(), usecases.InviteAccountCommand{
		Email: "user@example.com", FirstName: "A", LastName: "B",
	})

	// assert
	if err == nil {
		t.Fatalf("expected create error to propagate")
	}
}

func TestInviteAccount_IssueFails(t *testing.T) {
	// arrange
	invites := &invitesStub{issueErr: errors.New("token store")}
	uc := newInviteUC(newAccountsStub(), invites, &mailStub{})

	// act
	_, err := uc.Execute(context.Background(), usecases.InviteAccountCommand{
		Email: "user@example.com", FirstName: "A", LastName: "B",
	})

	// assert
	if err == nil {
		t.Fatalf("expected issue error to propagate")
	}
}

func TestInviteAccount_ResendForPendingHappyPath(t *testing.T) {
	// arrange
	pending := iamtest.NewAccount(t).WithStatus(iam.AccountStatusPendingPassword).Build(t)
	invites := &invitesStub{}
	mail := &mailStub{}
	uc := newInviteUC(newAccountsStub(), invites, mail)

	// act
	err := uc.ResendFor(context.Background(), pending)

	// assert
	test.NoErr(t, err)
	if len(invites.issued) != 1 || len(mail.invites) != 1 {
		t.Fatalf("expected re-issued invite and mail")
	}
	if mail.invites[0].Name != pending.Name().Full() {
		t.Fatalf("expected full recipient name %q, got %q", pending.Name().Full(), mail.invites[0].Name)
	}
}

func TestInviteAccount_ResendForActiveConflict(t *testing.T) {
	// arrange
	active := iamtest.NewAccount(t).Build(t)
	uc := newInviteUC(newAccountsStub(), &invitesStub{}, &mailStub{})

	// act
	err := uc.ResendFor(context.Background(), active)

	// assert
	test.ErrIs(t, err, shared.ErrConflict)
}
