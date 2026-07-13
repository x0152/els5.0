package iam

import (
	"context"

	"github.com/els/backend/internal/domain/shared/vo"
)

type AccountRepository interface {
	Create(ctx context.Context, a *Account) error
	Update(ctx context.Context, a *Account) error
	UpdatePicture(ctx context.Context, a *Account) (previousPictureURL string, err error)
	Delete(ctx context.Context, id AccountID) error
	GetByID(ctx context.Context, id AccountID) (*Account, error)
	GetByIDs(ctx context.Context, ids []AccountID) ([]*Account, error)
	GetByEmail(ctx context.Context, email vo.Email) (*Account, error)
	SearchByEmail(ctx context.Context, q string, limit int32) ([]*Account, error)
	ExistsEmail(ctx context.Context, email vo.Email) (bool, error)
}

type CredentialsRepository interface {
	Save(ctx context.Context, c *Credentials) error
	GetByAccountID(ctx context.Context, id AccountID) (*Credentials, error)
}

type AccountRoleRepository interface {
	GetByAccountID(ctx context.Context, id AccountID) (AccountRoleLink, error)
}
