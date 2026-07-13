package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/infrastructure/mappers"
	"github.com/els/backend/internal/infrastructure/postgres"
	"github.com/els/backend/internal/infrastructure/postgres/sqlc"
)

type AdministratorRepo struct {
	pool *pgxpool.Pool
	base *sqlc.Queries
}

func NewAdministratorRepo(pool *pgxpool.Pool) *AdministratorRepo {
	return &AdministratorRepo{pool: pool, base: sqlc.New(pool)}
}

func (r *AdministratorRepo) q(ctx context.Context) *sqlc.Queries {
	return postgres.QueriesFromContext(ctx, r.base)
}

func (r *AdministratorRepo) Create(ctx context.Context, a *admin.Administrator) error {
	if err := r.q(ctx).CreateAdministrator(ctx, mappers.AdministratorCreateParams(a)); err != nil {
		if postgres.IsUniqueViolation(err) {
			return fmt.Errorf("%w: account already bound to a role", shared.ErrConflict)
		}
		if postgres.IsForeignKeyViolation(err) {
			return fmt.Errorf("%w: client does not exist", shared.ErrValidation)
		}
		return fmt.Errorf("insert administrator: %w", err)
	}
	return nil
}

func (r *AdministratorRepo) Update(ctx context.Context, a *admin.Administrator) error {
	return postgres.RunTx(ctx, r.pool, func(ctx context.Context) error {
		q := r.q(ctx)
		accRows, err := q.UpdateAccount(ctx, mappers.AccountUpdateParams(a.Account()))
		if err != nil {
			if postgres.IsUniqueViolation(err) {
				return iam.ErrEmailTaken
			}
			return fmt.Errorf("update account: %w", err)
		}
		if accRows == 0 {
			return shared.ErrNotFound
		}
		rows, err := q.UpdateAdministrator(ctx, mappers.AdministratorUpdateParams(a))
		if err != nil {
			if postgres.IsForeignKeyViolation(err) {
				return fmt.Errorf("%w: client does not exist", shared.ErrValidation)
			}
			return fmt.Errorf("update administrator: %w", err)
		}
		if rows == 0 {
			return shared.ErrNotFound
		}
		return nil
	})
}

func (r *AdministratorRepo) Delete(ctx context.Context, id admin.ID) error {
	rows, err := r.q(ctx).DeleteAdministrator(ctx, sqlc.DeleteAdministratorParams{
		ID:        postgres.UUIDFromGoogle(id.UUID()),
		DeletedAt: postgres.TimestamptzFromTime(time.Now().UTC()),
	})
	if err != nil {
		return fmt.Errorf("delete administrator: %w", err)
	}
	if rows == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *AdministratorRepo) GetByID(ctx context.Context, id admin.ID) (*admin.Administrator, error) {
	q := r.q(ctx)
	row, err := q.GetAdministratorByID(ctx, postgres.UUIDFromGoogle(id.UUID()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("select administrator by id: %w", err)
	}
	acc, err := r.loadAccount(ctx, q, row.AccountID)
	if err != nil {
		return nil, err
	}
	return mappers.AdministratorFromSQLC(row, acc)
}

func (r *AdministratorRepo) GetByAccountID(ctx context.Context, id iam.AccountID) (*admin.Administrator, error) {
	q := r.q(ctx)
	row, err := q.GetAdministratorByAccountID(ctx, postgres.UUIDFromGoogle(id.UUID()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("select administrator by account id: %w", err)
	}
	acc, err := r.loadAccount(ctx, q, row.AccountID)
	if err != nil {
		return nil, err
	}
	return mappers.AdministratorFromSQLC(row, acc)
}

func (r *AdministratorRepo) List(ctx context.Context, filter admin.Filter, limit, offset int32) ([]*admin.Administrator, int64, error) {
	if filter.IsDeny() {
		return nil, 0, nil
	}
	q := r.q(ctx)
	rows, err := q.ListAdministrators(ctx, sqlc.ListAdministratorsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("list administrators: %w", err)
	}
	total, err := q.CountAdministrators(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count administrators: %w", err)
	}
	if len(rows) == 0 {
		return nil, total, nil
	}
	accountIDs := make([]pgtype.UUID, 0, len(rows))
	for _, row := range rows {
		accountIDs = append(accountIDs, row.AccountID)
	}
	accRows, err := q.GetAccountsByIDs(ctx, accountIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("select accounts: %w", err)
	}
	accByID := make(map[string]*iam.Account, len(accRows))
	for _, ar := range accRows {
		acc, err := mappers.AccountFromSQLC(ar)
		if err != nil {
			return nil, 0, err
		}
		accByID[acc.ID().String()] = acc
	}
	out := make([]*admin.Administrator, 0, len(rows))
	for _, row := range rows {
		accKey := postgres.UUIDToGoogle(row.AccountID).String()
		acc, ok := accByID[accKey]
		if !ok {
			return nil, 0, fmt.Errorf("account %s not found for administrator %s", accKey, postgres.UUIDToGoogle(row.ID))
		}
		a, err := mappers.AdministratorFromSQLC(row, acc)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, nil
}

func (r *AdministratorRepo) Count(ctx context.Context) (int64, error) {
	n, err := r.q(ctx).CountAdministrators(ctx)
	if err != nil {
		return 0, fmt.Errorf("count administrators: %w", err)
	}
	return n, nil
}

func (r *AdministratorRepo) loadAccount(ctx context.Context, q *sqlc.Queries, id pgtype.UUID) (*iam.Account, error) {
	accRow, err := q.GetAccountByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("select account: %w", err)
	}
	return mappers.AccountFromSQLC(accRow)
}
