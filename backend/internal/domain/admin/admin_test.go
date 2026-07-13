package admin_test

import (
	"testing"
	"time"

	admindom "github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestNewAdministrator(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(p *admindom.NewAdministratorParams)
		wantErr error
	}{
		{name: "ok", wantErr: nil},
		{name: "zero_id", mutate: func(p *admindom.NewAdministratorParams) {
			p.ID = admindom.ID{}
		}, wantErr: shared.ErrValidation},
		{name: "nil_account", mutate: func(p *admindom.NewAdministratorParams) {
			p.Account = nil
		}, wantErr: shared.ErrValidation},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts, err := vo.NewTimestamps(test.FixedTime, test.FixedTime)
			test.Must(t, err)
			p := admindom.NewAdministratorParams{
				ID:         admindom.NewID(),
				Account:    iamtest.NewAccount(t).Build(t),
				Timestamps: ts,
			}
			if tc.mutate != nil {
				tc.mutate(&p)
			}

			_, err = admindom.NewAdministrator(p)

			if tc.wantErr == nil {
				test.NoErr(t, err)
				return
			}
			test.ErrIs(t, err, tc.wantErr)
		})
	}
}

func TestNewAdministratorNow(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		acc := iamtest.NewAccount(t).Build(t)
		before := time.Now()

		a, err := admindom.NewAdministratorNow(admindom.NewID(), acc)
		after := time.Now()

		test.NoErr(t, err)
		if a.CreatedAt().Before(before) || a.CreatedAt().After(after) {
			t.Errorf("expected CreatedAt within [%v;%v], got %v", before, after, a.CreatedAt())
		}
	})

	t.Run("nil_account_returns_validation", func(t *testing.T) {
		_, err := admindom.NewAdministratorNow(admindom.NewID(), nil)
		test.ErrIs(t, err, shared.ErrValidation)
	})

	t.Run("zero_administrator_id_returns_validation", func(t *testing.T) {
		_, err := admindom.NewAdministratorNow(admindom.ID{}, iamtest.NewAccount(t).Build(t))
		test.ErrIs(t, err, shared.ErrValidation)
	})
}
