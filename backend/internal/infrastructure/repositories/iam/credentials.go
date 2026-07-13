package iam

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/infrastructure/mappers"
	"github.com/els/backend/internal/infrastructure/postgres"
	"github.com/els/backend/internal/infrastructure/postgres/sqlc"
	"github.com/els/backend/internal/utils/timex"
)

type CredentialsRepo struct {
	q *sqlc.Queries
}

func NewCredentialsRepo(pool *pgxpool.Pool) *CredentialsRepo {
	return &CredentialsRepo{q: sqlc.New(pool)}
}

func (r *CredentialsRepo) Save(ctx context.Context, c *iam.Credentials) error {
	err := r.q.UpsertCredentials(ctx, sqlc.UpsertCredentialsParams{
		AccountID:    postgres.UUIDFromGoogle(c.AccountID().UUID()),
		PasswordHash: c.Hash().String(),
		UpdatedAt:    postgres.TimestamptzFromTime(timex.Now()),
	})
	if err != nil {
		return fmt.Errorf("upsert credentials: %w", err)
	}
	return nil
}

func (r *CredentialsRepo) GetByAccountID(ctx context.Context, id iam.AccountID) (*iam.Credentials, error) {
	row, err := r.q.GetCredentialsByAccountID(ctx, postgres.UUIDFromGoogle(id.UUID()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("select credentials: %w", err)
	}
	return mappers.CredentialsFromSQLC(row)
}
