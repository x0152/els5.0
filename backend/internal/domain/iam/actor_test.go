package iam_test

import (
	"testing"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
)

func TestNewActor(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	entity := vo.NewID()

	cases := []struct {
		name    string
		mut     func(*iam.AccountRoleLink, **iam.Account)
		wantErr bool
	}{
		{name: "ok", mut: func(*iam.AccountRoleLink, **iam.Account) {}},
		{name: "nil_account", mut: func(_ *iam.AccountRoleLink, a **iam.Account) { *a = nil }, wantErr: true},
		{name: "invalid_role", mut: func(l *iam.AccountRoleLink, _ **iam.Account) { l.Role = "bogus" }, wantErr: true},
		{name: "zero_entity", mut: func(l *iam.AccountRoleLink, _ **iam.Account) { l.EntityID = vo.ID{} }, wantErr: true},
		{name: "account_id_mismatch", mut: func(l *iam.AccountRoleLink, _ **iam.Account) { l.AccountID = iam.NewAccountID() }, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := acc
			link := iam.AccountRoleLink{AccountID: acc.ID(), Role: iam.RoleExpert, EntityID: entity}
			tc.mut(&link, &a)

			actor, err := iam.NewActor(a, link)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			test.NoErr(t, err)
			if actor.AccountID() != acc.ID() {
				t.Fatalf("expected actor account id %s", acc.ID())
			}
			if actor.Role() != iam.RoleExpert {
				t.Fatalf("expected expert role")
			}
			if actor.IsGlobalAdmin() {
				t.Fatalf("non-admin must not be global admin")
			}
		})
	}
}

func TestActor_GlobalAdminFlag(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	link := iam.AccountRoleLink{AccountID: acc.ID(), Role: iam.RoleAdmin, EntityID: vo.NewID()}

	// act
	actor, err := iam.NewActor(acc, link)

	// assert
	test.NoErr(t, err)
	if !actor.IsGlobalAdmin() || !actor.Is(iam.RoleAdmin) {
		t.Fatalf("admin actor must be flagged as global admin")
	}
}

func TestRequireGlobalAdmin(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	admin, err := iam.NewActor(acc, iam.AccountRoleLink{AccountID: acc.ID(), Role: iam.RoleAdmin, EntityID: vo.NewID()})
	test.NoErr(t, err)
	expert, err := iam.NewActor(acc, iam.AccountRoleLink{AccountID: acc.ID(), Role: iam.RoleExpert, EntityID: vo.NewID()})
	test.NoErr(t, err)

	// act + assert
	test.NoErr(t, iam.RequireGlobalAdmin(admin))
	test.ErrIs(t, iam.RequireGlobalAdmin(expert), shared.ErrForbidden)
	test.ErrIs(t, iam.RequireGlobalAdmin(nil), shared.ErrForbidden)
}

func TestRequireRoleHelpers(t *testing.T) {
	// arrange
	acc, err := iam.NewAccount(validParams())
	test.NoErr(t, err)
	expert, _ := iam.NewActor(acc, iam.AccountRoleLink{AccountID: acc.ID(), Role: iam.RoleExpert, EntityID: vo.NewID()})
	customer, _ := iam.NewActor(acc, iam.AccountRoleLink{AccountID: acc.ID(), Role: iam.RoleCustomer, EntityID: vo.NewID()})
	admin, _ := iam.NewActor(acc, iam.AccountRoleLink{AccountID: acc.ID(), Role: iam.RoleAdmin, EntityID: vo.NewID()})

	// act + assert
	test.NoErr(t, iam.RequireExpert(expert))
	test.ErrIs(t, iam.RequireExpert(customer), shared.ErrForbidden)
	test.NoErr(t, iam.RequireCustomer(customer))
	test.ErrIs(t, iam.RequireCustomer(expert), shared.ErrForbidden)
	test.NoErr(t, iam.RequireAdmin(admin))
	test.ErrIs(t, iam.RequireAdmin(expert), shared.ErrForbidden)
	test.ErrIs(t, iam.RequireExpert(nil), shared.ErrForbidden)
}
