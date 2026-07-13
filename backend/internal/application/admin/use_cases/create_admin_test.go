package usecases_test

import (
	"context"
	"errors"
	"testing"

	usecases "github.com/els/backend/internal/application/admin/use_cases"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func validCreateAdminCmd() usecases.CreateAdminCommand {
	return usecases.CreateAdminCommand{Email: "new@example.com", FirstName: "New", LastName: "Admin"}
}

func TestCreateAdmin_Forbidden(t *testing.T) {
	// arrange
	repo := &adminRepoStub{}
	inv := &accountInviterStub{}
	uc := usecases.NewCreateAdminUseCase(repo, inv)

	// act
	_, err := uc.Execute(context.Background(), nil, validCreateAdminCmd())

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
	if len(inv.calls) != 0 || len(repo.createCalls) != 0 {
		t.Errorf("expected no side effects on forbidden")
	}
}

func TestCreateAdmin_InviteFails(t *testing.T) {
	// arrange
	boom := errors.New("invite failed")
	repo := &adminRepoStub{}
	inv := &accountInviterStub{err: boom}
	uc := usecases.NewCreateAdminUseCase(repo, inv)

	// act
	_, err := uc.Execute(context.Background(), iamtest.Admin(t), validCreateAdminCmd())

	// assert
	if !errors.Is(err, boom) {
		t.Errorf("expected invite error to propagate, got %v", err)
	}
	if len(repo.createCalls) != 0 {
		t.Errorf("expected no Create when invite failed")
	}
}

func TestCreateAdmin_RepoCreateFails(t *testing.T) {
	// arrange
	boom := errors.New("create failed")
	acc := iamtest.NewAccount(t).Build(t)
	repo := &adminRepoStub{createErr: boom}
	inv := &accountInviterStub{reply: acc}
	uc := usecases.NewCreateAdminUseCase(repo, inv)

	// act
	_, err := uc.Execute(context.Background(), iamtest.Admin(t), validCreateAdminCmd())

	// assert
	if !errors.Is(err, boom) {
		t.Errorf("expected repo error to propagate, got %v", err)
	}
}

func TestCreateAdmin_OK(t *testing.T) {
	// arrange
	acc := iamtest.NewAccount(t).Build(t)
	repo := &adminRepoStub{}
	inv := &accountInviterStub{reply: acc}
	uc := usecases.NewCreateAdminUseCase(repo, inv)
	cmd := validCreateAdminCmd()

	// act
	res, err := uc.Execute(context.Background(), iamtest.Admin(t), cmd)

	// assert
	test.NoErr(t, err)
	if len(inv.calls) != 1 || inv.calls[0].Email != cmd.Email {
		t.Errorf("expected one invite call with the same command, got %+v", inv.calls)
	}
	if len(repo.createCalls) != 1 {
		t.Fatalf("expected one Create call, got %d", len(repo.createCalls))
	}
	if res.Admin == nil || res.Admin.AccountID() != acc.ID() {
		t.Errorf("expected result admin bound to invited account")
	}
}
