package iam

import (
	"testing"

	iamdom "github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
)

type AccountBuilder struct {
	id         iamdom.AccountID
	email      string
	first      string
	last       string
	status     iamdom.AccountStatus
	pictureURL string
}

func NewAccount(t *testing.T) *AccountBuilder {
	t.Helper()
	return &AccountBuilder{
		id:     iamdom.NewAccountID(),
		email:  "user-" + iamdom.NewAccountID().String() + "@example.com",
		first:  "John",
		last:   "Doe",
		status: iamdom.AccountStatusActive,
	}
}

func (b *AccountBuilder) WithID(id iamdom.AccountID) *AccountBuilder { b.id = id; return b }
func (b *AccountBuilder) WithEmail(e string) *AccountBuilder         { b.email = e; return b }
func (b *AccountBuilder) WithName(first, last string) *AccountBuilder {
	b.first = first
	b.last = last
	return b
}
func (b *AccountBuilder) WithStatus(s iamdom.AccountStatus) *AccountBuilder { b.status = s; return b }
func (b *AccountBuilder) WithPictureURL(u string) *AccountBuilder           { b.pictureURL = u; return b }

func (b *AccountBuilder) Build(t *testing.T) *iamdom.Account {
	t.Helper()
	ts, err := vo.NewTimestamps(test.FixedTime, test.FixedTime)
	test.Must(t, err)
	acc, err := iamdom.NewAccount(iamdom.NewAccountParams{
		ID:         b.id,
		Email:      b.email,
		FirstName:  b.first,
		LastName:   b.last,
		PictureURL: b.pictureURL,
		Status:     b.status,
		Timestamps: ts,
	})
	test.Must(t, err)
	return acc
}
