package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestLoginConfirm_HappyPath(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).Build(t)
	accs := newAccountsStub().put(acc)
	roles := newRolesStub()
	roles.put(acc.ID(), iam.RoleExpert)
	invites := &invitesStub{
		consumeOK: ports.InviteToken{Purpose: ports.InviteTokenMagicLogin, AccountID: acc.ID().String()},
	}
	sess := &sessionsStub{createReply: "tok"}
	clk := test.FrozenClock()
	uc := usecases.NewLoginConfirmUseCase(accs, roles, invites, sess, time.Hour, clk)

	// act
	res, err := uc.Execute(context.Background(), usecases.LoginConfirmCommand{Token: "magic"})

	// assert
	test.NoErr(t, err)
	if res.Token != "tok" || res.Account == nil {
		t.Fatalf("unexpected result %+v", res)
	}
	if !res.ExpiresAt.Equal(clk.Now().Add(time.Hour)) {
		t.Fatalf("expected expires_at=%v, got %v", clk.Now().Add(time.Hour), res.ExpiresAt)
	}
	if len(sess.createCalls) != 1 || sess.createCalls[0].Subject.AccountID != acc.ID().String() {
		t.Fatalf("expected session created for account, got %+v", sess.createCalls)
	}
}

func TestLoginConfirm_TokenConsumeFails(t *testing.T) {
	// arrange
	invites := &invitesStub{consumeErr: errors.New("expired")}
	uc := usecases.NewLoginConfirmUseCase(newAccountsStub(), newRolesStub(), invites, &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginConfirmCommand{Token: "x"})

	// assert
	if err == nil {
		t.Fatalf("expected error from token consume")
	}
}

func TestLoginConfirm_WrongPurposeUnauthorized(t *testing.T) {
	// arrange
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenSetPassword, AccountID: "x"}}
	uc := usecases.NewLoginConfirmUseCase(newAccountsStub(), newRolesStub(), invites, &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginConfirmCommand{Token: "x"})

	// assert
	test.ErrIs(t, err, shared.ErrUnauthorized)
}

func TestLoginConfirm_AccountNotFoundUnauthorized(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).Build(t)
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenMagicLogin, AccountID: acc.ID().String()}}
	accs := newAccountsStub()
	uc := usecases.NewLoginConfirmUseCase(accs, newRolesStub(), invites, &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginConfirmCommand{Token: "x"})

	// assert
	test.ErrIs(t, err, shared.ErrUnauthorized)
}

func TestLoginConfirm_BlockedAccountForbidden(t *testing.T) {
	// arrange
	blocked := iamtest.NewAccount(t).WithStatus(iam.AccountStatusBlocked).Build(t)
	accs := newAccountsStub().put(blocked)
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenMagicLogin, AccountID: blocked.ID().String()}}
	uc := usecases.NewLoginConfirmUseCase(accs, newRolesStub(), invites, &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginConfirmCommand{Token: "x"})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestLoginConfirm_MissingRoleForbidden(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).Build(t)
	accs := newAccountsStub().put(acc)
	invites := &invitesStub{consumeOK: ports.InviteToken{Purpose: ports.InviteTokenMagicLogin, AccountID: acc.ID().String()}}
	uc := usecases.NewLoginConfirmUseCase(accs, newRolesStub(), invites, &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), usecases.LoginConfirmCommand{Token: "x"})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
}
