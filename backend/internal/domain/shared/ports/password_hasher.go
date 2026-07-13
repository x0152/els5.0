package ports

import "github.com/els/backend/internal/domain/shared/vo"

type PasswordHasher interface {
	Hash(plain string) (vo.PasswordHash, error)
	Verify(hash vo.PasswordHash, plain string) error
}
