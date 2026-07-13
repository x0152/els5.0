package iam_test

import (
	"errors"
	"testing"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
)

type fakeHasher struct{ verifyErr error }

func (h fakeHasher) Hash(_ string) (vo.PasswordHash, error)   { return vo.PasswordHash{}, nil }
func (h fakeHasher) Verify(_ vo.PasswordHash, _ string) error { return h.verifyErr }

func TestNewCredentials(t *testing.T) {
	id := iam.NewAccountID()
	hash, err := vo.NewPasswordHash("h")
	test.NoErr(t, err)

	cases := []struct {
		name    string
		id      iam.AccountID
		hash    vo.PasswordHash
		wantErr bool
	}{
		{name: "ok", id: id, hash: hash},
		{name: "zero_id", hash: hash, wantErr: true},
		{name: "empty_hash", id: id, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cred, err := iam.NewCredentials(tc.id, tc.hash)
			if tc.wantErr {
				test.ErrIs(t, err, shared.ErrValidation)
				return
			}
			test.NoErr(t, err)
			if cred.AccountID() != tc.id {
				t.Fatalf("expected account id %s, got %s", tc.id, cred.AccountID())
			}
		})
	}
}

func TestCredentials_Verify(t *testing.T) {
	id := iam.NewAccountID()
	hash, _ := vo.NewPasswordHash("h")
	cred, err := iam.NewCredentials(id, hash)
	test.NoErr(t, err)

	if err := cred.Verify("plain", fakeHasher{}); err != nil {
		t.Fatalf("expected verify ok, got %v", err)
	}
	err = cred.Verify("plain", fakeHasher{verifyErr: errors.New("mismatch")})
	if !errors.Is(err, iam.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
