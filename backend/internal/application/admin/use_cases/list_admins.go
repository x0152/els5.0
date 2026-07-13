package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
)

type ListAdminsUseCase struct {
	admins admin.Repository
}

func NewListAdminsUseCase(admins admin.Repository) *ListAdminsUseCase {
	return &ListAdminsUseCase{admins: admins}
}

type ListAdminsQuery struct {
	Limit  int32
	Offset int32
}

type ListAdminsResult struct {
	Admins []*admin.Administrator
	Total  int64
	Limit  int32
	Offset int32
}

func (uc *ListAdminsUseCase) Execute(
	ctx context.Context,
	actor *iam.Actor,
	q ListAdminsQuery,
) (ListAdminsResult, error) {
	// 1. Only a global admin can list admins.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return ListAdminsResult{}, err
	}

	// 2. Normalize pagination.
	limit, offset := normalizePage(q.Limit, q.Offset)

	// 3. Load the page + total.
	items, total, err := uc.admins.List(ctx, admin.VisibilityFor(actor), limit, offset)
	if err != nil {
		return ListAdminsResult{}, err
	}

	return ListAdminsResult{Admins: items, Total: total, Limit: limit, Offset: offset}, nil
}

func normalizePage(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
