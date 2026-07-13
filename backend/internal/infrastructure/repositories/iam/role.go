package iam

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/infrastructure/postgres"
	"github.com/els/backend/internal/infrastructure/postgres/sqlc"
)

type AccountRoleRepo struct {
	q *sqlc.Queries
}

func NewAccountRoleRepo(pool *pgxpool.Pool) *AccountRoleRepo {
	return &AccountRoleRepo{q: sqlc.New(pool)}
}

func (r *AccountRoleRepo) GetByAccountID(ctx context.Context, id iam.AccountID) (iam.AccountRoleLink, error) {
	row, err := r.q.GetAccountRoleByAccount(ctx, postgres.UUIDFromGoogle(id.UUID()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return iam.AccountRoleLink{}, shared.ErrNotFound
		}
		return iam.AccountRoleLink{}, fmt.Errorf("select account role: %w", err)
	}
	role, err := iam.ParseRole(row.Role)
	if err != nil {
		return iam.AccountRoleLink{}, fmt.Errorf("account %s: %w", id, err)
	}
	return iam.AccountRoleLink{
		AccountID: iam.AccountID{ID: vo.IDFromUUID(postgres.UUIDToGoogle(row.AccountID))},
		Role:      role,
		EntityID:  vo.IDFromUUID(postgres.UUIDToGoogle(row.EntityID)),
	}, nil
}
