package iam

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/infrastructure/mappers"
	"github.com/els/backend/internal/infrastructure/postgres"
	"github.com/els/backend/internal/infrastructure/postgres/sqlc"
)

type AccountRepo struct {
	base *sqlc.Queries
}

func NewAccountRepo(pool *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{base: sqlc.New(pool)}
}

func (r *AccountRepo) q(ctx context.Context) *sqlc.Queries {
	return postgres.QueriesFromContext(ctx, r.base)
}

func (r *AccountRepo) Create(ctx context.Context, a *iam.Account) error {
	if err := r.q(ctx).CreateAccount(ctx, mappers.AccountCreateParams(a)); err != nil {
		if postgres.IsUniqueViolation(err) {
			return iam.ErrEmailTaken
		}
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

func (r *AccountRepo) Update(ctx context.Context, a *iam.Account) error {
	rows, err := r.q(ctx).UpdateAccount(ctx, mappers.AccountUpdateParams(a))
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return iam.ErrEmailTaken
		}
		return fmt.Errorf("update account: %w", err)
	}
	if rows == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *AccountRepo) UpdatePicture(ctx context.Context, a *iam.Account) (string, error) {
	previous, err := r.q(ctx).UpdateAccountPicture(ctx, mappers.AccountUpdatePictureParams(a))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", shared.ErrNotFound
		}
		return "", fmt.Errorf("update account picture: %w", err)
	}
	return previous, nil
}

func (r *AccountRepo) Delete(ctx context.Context, id iam.AccountID) error {
	rows, err := r.q(ctx).DeleteAccount(ctx, postgres.UUIDFromGoogle(id.UUID()))
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if rows == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *AccountRepo) GetByID(ctx context.Context, id iam.AccountID) (*iam.Account, error) {
	row, err := r.q(ctx).GetAccountByID(ctx, postgres.UUIDFromGoogle(id.UUID()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("select account by id: %w", err)
	}
	return mappers.AccountFromSQLC(row)
}

func (r *AccountRepo) GetByEmail(ctx context.Context, email vo.Email) (*iam.Account, error) {
	row, err := r.q(ctx).GetAccountByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("select account by email: %w", err)
	}
	return mappers.AccountFromSQLC(row)
}

func (r *AccountRepo) ExistsEmail(ctx context.Context, email vo.Email) (bool, error) {
	exists, err := r.q(ctx).ExistsAccountEmail(ctx, email.String())
	if err != nil {
		return false, fmt.Errorf("exists email: %w", err)
	}
	return exists, nil
}

func (r *AccountRepo) GetByIDs(ctx context.Context, ids []iam.AccountID) ([]*iam.Account, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	pgIDs := make([]pgtype.UUID, 0, len(ids))
	for _, id := range ids {
		pgIDs = append(pgIDs, postgres.UUIDFromGoogle(id.UUID()))
	}
	rows, err := r.q(ctx).GetAccountsByIDs(ctx, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("select accounts by ids: %w", err)
	}
	out := make([]*iam.Account, 0, len(rows))
	for _, row := range rows {
		a, err := mappers.AccountFromSQLC(row)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *AccountRepo) SearchByEmail(ctx context.Context, q string, limit int32) ([]*iam.Account, error) {
	if limit <= 0 {
		limit = 20
	}
	needle := "%" + strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(q), `\`, `\\`), `%`, `\%`), `_`, `\_`) + "%"
	rows, err := r.q(ctx).SearchAccountsByEmail(ctx, sqlc.SearchAccountsByEmailParams{Query: needle, Lim: limit})
	if err != nil {
		return nil, fmt.Errorf("search accounts by email: %w", err)
	}
	out := make([]*iam.Account, 0, len(rows))
	for _, row := range rows {
		a, err := mappers.AccountFromSQLC(row)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}
