package mappers

import (
	"fmt"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/infrastructure/postgres"
	"github.com/els/backend/internal/infrastructure/postgres/sqlc"
)

func AccountFromSQLC(r sqlc.Account) (*iam.Account, error) {
	id := iam.AccountID{ID: vo.IDFromUUID(postgres.UUIDToGoogle(r.ID))}
	timestamps, err := vo.NewTimestamps(
		postgres.TimestamptzToTime(r.CreatedAt),
		postgres.TimestamptzToTime(r.UpdatedAt),
	)
	if err != nil {
		return nil, fmt.Errorf("account %s timestamps: %w", id, err)
	}
	status, err := iam.ParseAccountStatus(r.Status)
	if err != nil {
		return nil, fmt.Errorf("account %s status: %w", id, err)
	}
	return iam.NewAccount(iam.NewAccountParams{
		ID:           id,
		Email:        r.Email,
		FirstName:    r.FirstName,
		LastName:     r.LastName,
		PictureURL:   r.PictureUrl,
		EnglishLevel: r.EnglishLevel,
		AboutMe:      r.AboutMe,
		Status:       status,
		Timestamps:   timestamps,
	})
}

func AccountCreateParams(a *iam.Account) sqlc.CreateAccountParams {
	return sqlc.CreateAccountParams{
		ID:           postgres.UUIDFromGoogle(a.ID().UUID()),
		Email:        a.Email().String(),
		FirstName:    a.Name().First(),
		LastName:     a.Name().Last(),
		PictureUrl:   a.PictureURL(),
		EnglishLevel: a.EnglishLevel(),
		AboutMe:      a.AboutMe(),
		Status:       a.Status().String(),
		CreatedAt:    postgres.TimestamptzFromTime(a.CreatedAt()),
		UpdatedAt:    postgres.TimestamptzFromTime(a.UpdatedAt()),
	}
}

func AccountUpdateParams(a *iam.Account) sqlc.UpdateAccountParams {
	return sqlc.UpdateAccountParams{
		ID:           postgres.UUIDFromGoogle(a.ID().UUID()),
		Email:        a.Email().String(),
		FirstName:    a.Name().First(),
		LastName:     a.Name().Last(),
		Status:       a.Status().String(),
		EnglishLevel: a.EnglishLevel(),
		AboutMe:      a.AboutMe(),
		UpdatedAt:    postgres.TimestamptzFromTime(a.UpdatedAt()),
	}
}

func AccountUpdatePictureParams(a *iam.Account) sqlc.UpdateAccountPictureParams {
	return sqlc.UpdateAccountPictureParams{
		ID:         postgres.UUIDFromGoogle(a.ID().UUID()),
		PictureUrl: a.PictureURL(),
		UpdatedAt:  postgres.TimestamptzFromTime(a.UpdatedAt()),
	}
}

func CredentialsFromSQLC(r sqlc.Credential) (*iam.Credentials, error) {
	hash, err := vo.NewPasswordHash(r.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("credentials %s: %w", postgres.UUIDToGoogle(r.AccountID), err)
	}
	id := iam.AccountID{ID: vo.IDFromUUID(postgres.UUIDToGoogle(r.AccountID))}
	return iam.HydrateCredentials(id, hash), nil
}
