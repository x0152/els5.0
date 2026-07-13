package usecases_test

import (
	"context"
	"testing"
	"time"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestImpersonate_HappyPath(t *testing.T) {
	// arrange
	admin := iamtest.Admin(t)
	target := iamtest.NewAccount(t).WithEmail("expert@example.com").Build(t)
	accs := newAccountsStub().put(admin.Account()).put(target)
	roles := newRolesStub()
	roles.put(target.ID(), iam.RoleExpert)
	sess := &sessionsStub{createReply: "imp-tok"}
	uc := usecases.NewImpersonateUseCase(true, accs, roles, sess, time.Hour, test.FrozenClock())

	// act
	res, err := uc.Execute(context.Background(), admin, usecases.ImpersonateCommand{
		TargetAccountID: target.ID(),
	})

	// assert
	test.NoErr(t, err)
	if res.Token != "imp-tok" || res.Account == nil || res.Account.ID() != target.ID() {
		t.Fatalf("unexpected result %+v", res)
	}
	if len(sess.createCalls) != 1 || sess.createCalls[0].Subject.AccountID != target.ID().String() {
		t.Fatalf("expected session created for target, got %+v", sess.createCalls)
	}
	if sess.createCalls[0].Subject.Role != iam.RoleExpert.String() {
		t.Fatalf("expected role=%s, got %s", iam.RoleExpert, sess.createCalls[0].Subject.Role)
	}
}

func TestImpersonate_DisabledForbidden(t *testing.T) {
	// arrange
	admin := iamtest.Admin(t)
	target := iamtest.NewAccount(t).Build(t)
	accs := newAccountsStub().put(admin.Account()).put(target)
	roles := newRolesStub()
	roles.put(target.ID(), iam.RoleExpert)
	uc := usecases.NewImpersonateUseCase(false, accs, roles, &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), admin, usecases.ImpersonateCommand{
		TargetAccountID: target.ID(),
	})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestImpersonate_NonGlobalAdminForbidden(t *testing.T) {
	// arrange
	expert := iamtest.NewAccount(t).Build(t)
	expertActor := iamtest.ExpertFor(t, expert, expert.ID().ID)
	target := iamtest.NewAccount(t).WithEmail("other@example.com").Build(t)
	accs := newAccountsStub().put(expert).put(target)
	roles := newRolesStub()
	roles.put(target.ID(), iam.RoleExpert)
	uc := usecases.NewImpersonateUseCase(true, accs, roles, &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), expertActor, usecases.ImpersonateCommand{
		TargetAccountID: target.ID(),
	})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestImpersonate_SelfRejected(t *testing.T) {
	// arrange
	admin := iamtest.Admin(t)
	accs := newAccountsStub().put(admin.Account())
	uc := usecases.NewImpersonateUseCase(true, accs, newRolesStub(), &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), admin, usecases.ImpersonateCommand{
		TargetAccountID: admin.AccountID(),
	})

	// assert
	test.ErrIs(t, err, shared.ErrValidation)
}

func TestImpersonate_TargetNotFound(t *testing.T) {
	// arrange
	admin := iamtest.Admin(t)
	missing := iamtest.NewAccount(t).Build(t)
	accs := newAccountsStub().put(admin.Account())
	uc := usecases.NewImpersonateUseCase(true, accs, newRolesStub(), &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), admin, usecases.ImpersonateCommand{
		TargetAccountID: missing.ID(),
	})

	// assert
	test.ErrIs(t, err, shared.ErrNotFound)
}

func TestImpersonate_BlockedTarget(t *testing.T) {
	// arrange
	admin := iamtest.Admin(t)
	blocked := iamtest.NewAccount(t).WithStatus(iam.AccountStatusBlocked).Build(t)
	accs := newAccountsStub().put(admin.Account()).put(blocked)
	uc := usecases.NewImpersonateUseCase(true, accs, newRolesStub(), &sessionsStub{}, time.Hour, test.FrozenClock())

	// act
	_, err := uc.Execute(context.Background(), admin, usecases.ImpersonateCommand{
		TargetAccountID: blocked.ID(),
	})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
}
