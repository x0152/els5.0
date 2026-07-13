package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/practice"
)

type GetProgressUseCase struct {
	progress practice.ProgressRepository
}

func NewGetProgressUseCase(progress practice.ProgressRepository) *GetProgressUseCase {
	return &GetProgressUseCase{progress: progress}
}

func (uc *GetProgressUseCase) Execute(ctx context.Context, actor *iam.Actor, kind practice.Kind, number int, variantKey string) (practice.Progress, error) {
	return uc.progress.Get(ctx, actor.AccountID().String(), kind, number, variantKey)
}

type SaveProgressUseCase struct {
	progress practice.ProgressRepository
}

func NewSaveProgressUseCase(progress practice.ProgressRepository) *SaveProgressUseCase {
	return &SaveProgressUseCase{progress: progress}
}

func (uc *SaveProgressUseCase) Execute(ctx context.Context, actor *iam.Actor, kind practice.Kind, number int, variantKey string, p practice.Progress) error {
	return uc.progress.Save(ctx, actor.AccountID().String(), kind, number, variantKey, p)
}

type ResetProgressUseCase struct {
	progress practice.ProgressRepository
}

func NewResetProgressUseCase(progress practice.ProgressRepository) *ResetProgressUseCase {
	return &ResetProgressUseCase{progress: progress}
}

func (uc *ResetProgressUseCase) Execute(ctx context.Context, actor *iam.Actor, kind practice.Kind, number int, variantKey string) error {
	return uc.progress.Delete(ctx, actor.AccountID().String(), kind, number, variantKey)
}
