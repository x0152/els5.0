package iam

import (
	"fmt"
	"sort"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

type AccountRoleLink struct {
	AccountID AccountID
	Role      Role
	EntityID  vo.ID
}

type Scope struct {
	IDs []vo.ID
}

type Actor struct {
	account       *Account
	role          Role
	entityID      vo.ID
	isGlobalAdmin bool
	scope         Scope
}

func NewActor(account *Account, link AccountRoleLink) (*Actor, error) {
	return NewActorWithScope(account, link, Scope{})
}

func NewActorWithScope(account *Account, link AccountRoleLink, scope Scope) (*Actor, error) {
	if account == nil {
		return nil, fmt.Errorf("actor.account: must not be nil")
	}
	if !link.Role.IsValid() {
		return nil, fmt.Errorf("actor.role: invalid %q", link.Role)
	}
	if link.EntityID.IsZero() {
		return nil, fmt.Errorf("actor.entity_id: must not be zero")
	}
	if link.AccountID != account.ID() {
		return nil, fmt.Errorf("actor.account_id mismatch: link %s vs account %s", link.AccountID, account.ID())
	}
	normalized := normalizeScopeIDs(scope.IDs)
	return &Actor{
		account:       account,
		role:          link.Role,
		entityID:      link.EntityID,
		isGlobalAdmin: link.Role == RoleAdmin && len(normalized) == 0,
		scope:         Scope{IDs: normalized},
	}, nil
}

func (a *Actor) Account() *Account    { return a.account }
func (a *Actor) AccountID() AccountID { return a.account.ID() }
func (a *Actor) Role() Role           { return a.role }
func (a *Actor) EntityID() vo.ID      { return a.entityID }
func (a *Actor) IsGlobalAdmin() bool  { return a.isGlobalAdmin }
func (a *Actor) Is(role Role) bool    { return a.role == role }
func (a *Actor) IsAdmin() bool        { return a.Is(RoleAdmin) }
func (a *Actor) IsCustomer() bool     { return a.Is(RoleCustomer) }
func (a *Actor) IsExpert() bool       { return a.Is(RoleExpert) }
func (a *Actor) Scope() Scope         { return a.scope }

func (a *Actor) ScopeIDs() []vo.ID {
	if a == nil || len(a.scope.IDs) == 0 {
		return nil
	}
	out := make([]vo.ID, len(a.scope.IDs))
	copy(out, a.scope.IDs)
	return out
}

func (a *Actor) HasScope() bool {
	return a != nil && len(a.scope.IDs) > 0
}

func (a *Actor) IsScopedTo(id vo.ID) bool {
	if a == nil || id.IsZero() {
		return false
	}
	for _, scopeID := range a.scope.IDs {
		if scopeID == id {
			return true
		}
	}
	return false
}

func RequireGlobalAdmin(a *Actor) error {
	if a == nil || !a.IsGlobalAdmin() {
		return fmt.Errorf("%w: requires global administrator", shared.ErrForbidden)
	}
	return nil
}

func RequireSelfOrGlobalAdmin(a *Actor, target AccountID) error {
	if a == nil {
		return fmt.Errorf("%w: requires actor", shared.ErrUnauthorized)
	}
	if a.AccountID() == target {
		return nil
	}
	return RequireGlobalAdmin(a)
}

func RequireExpert(a *Actor) error   { return requireRole(a, RoleExpert) }
func RequireCustomer(a *Actor) error { return requireRole(a, RoleCustomer) }
func RequireAdmin(a *Actor) error    { return requireRole(a, RoleAdmin) }

func requireRole(a *Actor, role Role) error {
	if a == nil || a.Role() != role {
		return fmt.Errorf("%w: requires role %s", shared.ErrForbidden, role)
	}
	return nil
}

func normalizeScopeIDs(in []vo.ID) []vo.ID {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[vo.ID]struct{}, len(in))
	out := make([]vo.ID, 0, len(in))
	for _, id := range in {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}
