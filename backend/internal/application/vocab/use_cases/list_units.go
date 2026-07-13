package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type ListUnitsUseCase struct {
	units vocab.Repository
}

func NewListUnitsUseCase(units vocab.Repository) *ListUnitsUseCase {
	return &ListUnitsUseCase{units: units}
}

type ListUnitsResult struct {
	Items  []vocab.Unit
	Total  int
	Limit  int
	Offset int
}

func (uc *ListUnitsUseCase) Execute(ctx context.Context, actor *iam.Actor, filter vocab.ListFilter) (ListUnitsResult, error) {
	if actor == nil {
		return ListUnitsResult{}, shared.ErrUnauthorized
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	items, total, err := uc.units.List(ctx, actor.AccountID().String(), filter)
	if err != nil {
		return ListUnitsResult{}, err
	}
	return ListUnitsResult{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset}, nil
}
