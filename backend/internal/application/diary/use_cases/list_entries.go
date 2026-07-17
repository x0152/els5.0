package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/domain/iam"
)

type ListEntriesQuery struct {
	Limit  int32
	Offset int32
}

type ListEntriesResult struct {
	Items []diary.Entry
	Total int64
}

type ListEntriesUseCase struct {
	repo diary.Repository
}

func NewListEntriesUseCase(repo diary.Repository) *ListEntriesUseCase {
	return &ListEntriesUseCase{repo: repo}
}

func (uc *ListEntriesUseCase) Execute(ctx context.Context, actor *iam.Actor, q ListEntriesQuery) (ListEntriesResult, error) {
	// 1. Load the page of entries for the account.
	items, total, err := uc.repo.List(ctx, actor.AccountID().String(), q.Limit, q.Offset)
	if err != nil {
		return ListEntriesResult{}, err
	}
	return ListEntriesResult{Items: items, Total: total}, nil
}
