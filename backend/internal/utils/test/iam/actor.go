package iam

import (
	"testing"

	iamdom "github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
)

func Admin(t *testing.T) *iamdom.Actor {
	t.Helper()
	acc := newActiveAccount(t, "admin@example.com", "Admin", "Root")
	return newActor(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleAdmin,
		EntityID:  acc.ID().ID,
	})
}

func Expert(t *testing.T, expertID vo.ID) *iamdom.Actor {
	t.Helper()
	acc := newActiveAccount(t, "expert@example.com", "Ex", "Pert")
	return newActor(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleExpert,
		EntityID:  expertID,
	})
}

func ExpertFor(t *testing.T, acc *iamdom.Account, expertID vo.ID) *iamdom.Actor {
	t.Helper()
	return newActor(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleExpert,
		EntityID:  expertID,
	})
}

func Customer(t *testing.T, customerID vo.ID) *iamdom.Actor {
	t.Helper()
	acc := newActiveAccount(t, "customer@example.com", "Cu", "Stomer")
	return newActor(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleCustomer,
		EntityID:  customerID,
	})
}

func CustomerFor(t *testing.T, acc *iamdom.Account, customerID vo.ID) *iamdom.Actor {
	t.Helper()
	return newActor(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleCustomer,
		EntityID:  customerID,
	})
}

func CustomerWithClient(t *testing.T, customerID, clientID vo.ID) *iamdom.Actor {
	t.Helper()
	acc := newActiveAccount(t, "customer@example.com", "Cu", "Stomer")
	return newActorWithScope(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleCustomer,
		EntityID:  customerID,
	}, iamdom.Scope{IDs: []vo.ID{clientID}})
}

func CustomerForWithClient(t *testing.T, acc *iamdom.Account, customerID, clientID vo.ID) *iamdom.Actor {
	t.Helper()
	return newActorWithScope(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleCustomer,
		EntityID:  customerID,
	}, iamdom.Scope{IDs: []vo.ID{clientID}})
}

func CustomerWithClients(t *testing.T, customerID vo.ID, clientIDs ...vo.ID) *iamdom.Actor {
	t.Helper()
	acc := newActiveAccount(t, "customer@example.com", "Cu", "Stomer")
	return newActorWithScope(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleCustomer,
		EntityID:  customerID,
	}, iamdom.Scope{IDs: clientIDs})
}

func AdminScopedToClient(t *testing.T, adminID, clientID vo.ID) *iamdom.Actor {
	t.Helper()
	acc := newActiveAccount(t, "admin@example.com", "Admin", "Scoped")
	return newActorWithScope(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleAdmin,
		EntityID:  adminID,
	}, iamdom.Scope{IDs: []vo.ID{clientID}})
}

func AdminScopedToClientFor(t *testing.T, acc *iamdom.Account, adminID, clientID vo.ID) *iamdom.Actor {
	t.Helper()
	return newActorWithScope(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleAdmin,
		EntityID:  adminID,
	}, iamdom.Scope{IDs: []vo.ID{clientID}})
}

func AdminScopedToClients(t *testing.T, adminID vo.ID, clientIDs ...vo.ID) *iamdom.Actor {
	t.Helper()
	acc := newActiveAccount(t, "admin@example.com", "Admin", "Scoped")
	return newActorWithScope(t, acc, iamdom.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      iamdom.RoleAdmin,
		EntityID:  adminID,
	}, iamdom.Scope{IDs: clientIDs})
}

func newActor(t *testing.T, acc *iamdom.Account, link iamdom.AccountRoleLink) *iamdom.Actor {
	t.Helper()
	a, err := iamdom.NewActor(acc, link)
	test.Must(t, err)
	return a
}

func newActorWithScope(t *testing.T, acc *iamdom.Account, link iamdom.AccountRoleLink, scope iamdom.Scope) *iamdom.Actor {
	t.Helper()
	a, err := iamdom.NewActorWithScope(acc, link, scope)
	test.Must(t, err)
	return a
}

func newActiveAccount(t *testing.T, email, first, last string) *iamdom.Account {
	t.Helper()
	ts, err := vo.NewTimestamps(test.FixedTime, test.FixedTime)
	test.Must(t, err)
	acc, err := iamdom.NewAccount(iamdom.NewAccountParams{
		ID:         iamdom.NewAccountID(),
		Email:      email,
		FirstName:  first,
		LastName:   last,
		Status:     iamdom.AccountStatusActive,
		Timestamps: ts,
	})
	test.Must(t, err)
	return acc
}
