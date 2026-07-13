package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/httpx"
)

const bearerPrefix = "Bearer "

type ClientScopeResolver interface {
	ResolveClientIDs(ctx context.Context, accountID iam.AccountID, role iam.Role) ([]vo.ID, error)
}

type Authenticator struct {
	sessions ports.SessionStore
	accounts iam.AccountRepository
	scopes   ClientScopeResolver
}

func New(sessions ports.SessionStore, accounts iam.AccountRepository) *Authenticator {
	return &Authenticator{sessions: sessions, accounts: accounts}
}

func NewWithClientScope(
	sessions ports.SessionStore,
	accounts iam.AccountRepository,
	scopes ClientScopeResolver,
) *Authenticator {
	return &Authenticator{sessions: sessions, accounts: accounts, scopes: scopes}
}

func (a *Authenticator) ExtractToken(authorization string) (string, error) {
	v := strings.TrimSpace(authorization)
	if len(v) <= len(bearerPrefix) || !strings.EqualFold(v[:len(bearerPrefix)], bearerPrefix) {
		return "", unauthorized("missing bearer token")
	}
	token := strings.TrimSpace(v[len(bearerPrefix):])
	if token == "" {
		return "", unauthorized("missing bearer token")
	}
	return token, nil
}

func (a *Authenticator) Authenticate(ctx context.Context, authorization string) (*iam.Actor, string, error) {
	token, err := a.ExtractToken(authorization)
	if err != nil {
		return nil, "", err
	}

	subject, err := a.sessions.Lookup(ctx, token)
	if err != nil {
		return nil, "", err
	}

	id, err := parseAccountID(subject.AccountID)
	if err != nil {
		return nil, "", shared.ErrUnauthorized
	}
	acc, err := a.accounts.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, "", shared.ErrUnauthorized
		}
		return nil, "", fmt.Errorf("load account: %w", err)
	}

	if err := acc.EnsureCanLogin(); err != nil {
		return nil, "", err
	}

	actor, err := a.actorFromSubject(ctx, acc, subject)
	if err != nil {
		return nil, "", shared.ErrUnauthorized
	}

	return actor, token, nil
}

func (a *Authenticator) actorFromSubject(ctx context.Context, acc *iam.Account, subject ports.SessionSubject) (*iam.Actor, error) {
	role, err := iam.ParseRole(subject.Role)
	if err != nil {
		return nil, err
	}
	entityID, err := vo.ParseID(subject.EntityID)
	if err != nil {
		return nil, err
	}
	link := iam.AccountRoleLink{
		AccountID: acc.ID(),
		Role:      role,
		EntityID:  entityID,
	}
	scope, err := a.resolveScope(ctx, acc.ID(), role)
	if err != nil {
		return nil, err
	}
	return iam.NewActorWithScope(acc, link, scope)
}

func (a *Authenticator) resolveScope(ctx context.Context, accountID iam.AccountID, role iam.Role) (iam.Scope, error) {
	if a.scopes == nil {
		return iam.Scope{}, nil
	}
	if role != iam.RoleCustomer && role != iam.RoleAdmin {
		return iam.Scope{}, nil
	}
	clientIDs, err := a.scopes.ResolveClientIDs(ctx, accountID, role)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return iam.Scope{}, nil
		}
		return iam.Scope{}, fmt.Errorf("resolve client scope: %w", err)
	}
	return iam.Scope{IDs: clientIDs}, nil
}

func parseAccountID(s string) (iam.AccountID, error) {
	id, err := vo.ParseID(s)
	if err != nil {
		return iam.AccountID{}, err
	}
	return iam.AccountID{ID: id}, nil
}

func unauthorized(msg string) error {
	return httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthorized, msg)
}
