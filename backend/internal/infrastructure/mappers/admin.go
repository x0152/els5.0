package mappers

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/infrastructure/postgres"
	"github.com/els/backend/internal/infrastructure/postgres/sqlc"
)

type administratorRow struct {
	ID        pgtype.UUID
	AccountID pgtype.UUID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func AdministratorFromSQLC(r sqlc.Administrator, account *iam.Account) (*admin.Administrator, error) {
	return administratorFromRow(administratorRow{
		ID:        r.ID,
		AccountID: r.AccountID,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}, account)
}

func administratorFromRow(r administratorRow, account *iam.Account) (*admin.Administrator, error) {
	id := admin.ID{ID: vo.IDFromUUID(postgres.UUIDToGoogle(r.ID))}
	timestamps, err := vo.NewTimestamps(
		postgres.TimestamptzToTime(r.CreatedAt),
		postgres.TimestamptzToTime(r.UpdatedAt),
	)
	if err != nil {
		return nil, fmt.Errorf("administrator %s timestamps: %w", postgres.UUIDToGoogle(r.ID), err)
	}
	return admin.NewAdministrator(admin.NewAdministratorParams{
		ID:         id,
		Account:    account,
		Timestamps: timestamps,
	})
}

func AdministratorCreateParams(a *admin.Administrator) sqlc.CreateAdministratorParams {
	return sqlc.CreateAdministratorParams{
		ID:        postgres.UUIDFromGoogle(a.ID().UUID()),
		AccountID: postgres.UUIDFromGoogle(a.AccountID().UUID()),
		CreatedAt: postgres.TimestamptzFromTime(a.CreatedAt()),
		UpdatedAt: postgres.TimestamptzFromTime(a.UpdatedAt()),
	}
}

func AdministratorUpdateParams(a *admin.Administrator) sqlc.UpdateAdministratorParams {
	return sqlc.UpdateAdministratorParams{
		ID:        postgres.UUIDFromGoogle(a.ID().UUID()),
		UpdatedAt: postgres.TimestamptzFromTime(a.UpdatedAt()),
	}
}
