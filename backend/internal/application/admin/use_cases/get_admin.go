package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
)

type GetAdminUseCase struct {
	admins admin.Repository
}

func NewGetAdminUseCase(admins admin.Repository) *GetAdminUseCase {
	return &GetAdminUseCase{admins: admins}
}

type GetAdminQuery struct {
	ID admin.ID
}

type GetAdminResult struct {
	Admin *admin.Administrator
}

func (uc *GetAdminUseCase) Execute(
	ctx context.Context,
	actor *iam.Actor,
	q GetAdminQuery,
) (GetAdminResult, error) {
	// 1. Only a global admin can view admin cards.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return GetAdminResult{}, err
	}

	// 2. Read the administrator from the repository.
	a, err := uc.admins.GetByID(ctx, q.ID)
	if err != nil {
		return GetAdminResult{}, err
	}

	return GetAdminResult{Admin: a}, nil
}
