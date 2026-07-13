package admin

import (
	"testing"
	"time"

	admindom "github.com/els/backend/internal/domain/admin"
	iamdom "github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

type Builder struct {
	id        admindom.ID
	account   *iamdom.Account
	createdAt time.Time
	updatedAt time.Time
}

func New(t *testing.T) *Builder {
	t.Helper()
	return &Builder{
		id:        admindom.NewID(),
		account:   iamtest.NewAccount(t).Build(t),
		createdAt: test.FixedTime,
		updatedAt: test.FixedTime,
	}
}

func (b *Builder) WithID(id admindom.ID) *Builder         { b.id = id; return b }
func (b *Builder) WithAccount(a *iamdom.Account) *Builder { b.account = a; return b }
func (b *Builder) WithCreatedAt(at time.Time) *Builder    { b.createdAt = at; return b }
func (b *Builder) WithUpdatedAt(at time.Time) *Builder    { b.updatedAt = at; return b }

func (b *Builder) Build(t *testing.T) *admindom.Administrator {
	t.Helper()
	ts, err := vo.NewTimestamps(b.createdAt, b.updatedAt)
	test.Must(t, err)
	a, err := admindom.NewAdministrator(admindom.NewAdministratorParams{
		ID:         b.id,
		Account:    b.account,
		Timestamps: ts,
	})
	test.Must(t, err)
	return a
}
