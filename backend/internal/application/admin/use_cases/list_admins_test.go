package usecases_test

import (
	"context"
	"errors"
	"testing"

	usecases "github.com/els/backend/internal/application/admin/use_cases"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
	admintest "github.com/els/backend/internal/utils/test/admin"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestListAdmins_Forbidden(t *testing.T) {
	// arrange
	repo := &adminRepoStub{}
	uc := usecases.NewListAdminsUseCase(repo)

	// act
	_, err := uc.Execute(context.Background(), nil, usecases.ListAdminsQuery{Limit: 10})

	// assert
	test.ErrIs(t, err, shared.ErrForbidden)
	if repo.listCalls != 0 {
		t.Errorf("expected no List calls on forbidden")
	}
}

func TestListAdmins_PaginationDefaults(t *testing.T) {
	cases := []struct {
		name       string
		limit      int32
		offset     int32
		wantLimit  int32
		wantOffset int32
	}{
		{name: "zero_limit_uses_50", limit: 0, offset: 0, wantLimit: 50, wantOffset: 0},
		{name: "limit_capped_at_200", limit: 999, offset: 0, wantLimit: 200, wantOffset: 0},
		{name: "negative_offset_clamped", limit: 10, offset: -3, wantLimit: 10, wantOffset: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			repo := &adminRepoStub{}
			uc := usecases.NewListAdminsUseCase(repo)

			// act
			res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ListAdminsQuery{
				Limit:  tc.limit,
				Offset: tc.offset,
			})

			// assert
			test.NoErr(t, err)
			if repo.listLimit != tc.wantLimit || repo.listOffset != tc.wantOffset {
				t.Errorf("expected repo (%d,%d), got (%d,%d)",
					tc.wantLimit, tc.wantOffset, repo.listLimit, repo.listOffset)
			}
			if res.Limit != tc.wantLimit || res.Offset != tc.wantOffset {
				t.Errorf("expected result (%d,%d), got (%d,%d)",
					tc.wantLimit, tc.wantOffset, res.Limit, res.Offset)
			}
		})
	}
}

func TestListAdmins_ReturnsRepoData(t *testing.T) {
	// arrange
	items := []*admin.Administrator{admintest.New(t).Build(t), admintest.New(t).Build(t)}
	repo := &adminRepoStub{listReply: items, listTotal: 7}
	uc := usecases.NewListAdminsUseCase(repo)

	// act
	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ListAdminsQuery{Limit: 10})

	// assert
	test.NoErr(t, err)
	if len(res.Admins) != 2 || res.Admins[0] != items[0] {
		t.Errorf("expected items returned as-is from repo")
	}
	if res.Total != 7 {
		t.Errorf("expected total=7, got %d", res.Total)
	}
	if repo.listFilter.IsDeny() {
		t.Errorf("expected non-deny filter for global admin")
	}
}

func TestListAdmins_RepoFails(t *testing.T) {
	// arrange
	boom := errors.New("list failed")
	repo := &adminRepoStub{listErr: boom}
	uc := usecases.NewListAdminsUseCase(repo)

	// act
	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.ListAdminsQuery{Limit: 10})

	// assert
	if !errors.Is(err, boom) {
		t.Errorf("expected list error to propagate, got %v", err)
	}
}
