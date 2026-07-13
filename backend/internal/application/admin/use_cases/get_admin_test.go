package usecases_test

import (
	"context"
	"testing"

	usecases "github.com/els/backend/internal/application/admin/use_cases"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
	admintest "github.com/els/backend/internal/utils/test/admin"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestGetAdmin_Forbidden(t *testing.T) {
	repo := &adminRepoStub{}
	uc := usecases.NewGetAdminUseCase(repo)

	_, err := uc.Execute(context.Background(), nil, usecases.GetAdminQuery{ID: admin.ID{ID: vo.NewID()}})

	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestGetAdmin_NotFound(t *testing.T) {
	repo := &adminRepoStub{
		getByID: func(_ context.Context, _ admin.ID) (*admin.Administrator, error) {
			return nil, shared.ErrNotFound
		},
	}
	uc := usecases.NewGetAdminUseCase(repo)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.GetAdminQuery{ID: admin.ID{ID: vo.NewID()}})

	test.ErrIs(t, err, shared.ErrNotFound)
}

func TestGetAdmin_OK(t *testing.T) {
	existing := admintest.New(t).Build(t)
	repo := &adminRepoStub{
		getByID: func(_ context.Context, _ admin.ID) (*admin.Administrator, error) { return existing, nil },
	}
	uc := usecases.NewGetAdminUseCase(repo)

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), usecases.GetAdminQuery{ID: existing.ID()})

	test.NoErr(t, err)
	if res.Admin != existing {
		t.Errorf("expected returned administrator to be the one from repo")
	}
}
