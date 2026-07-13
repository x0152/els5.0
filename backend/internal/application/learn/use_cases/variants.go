package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/practice"
)

// SourceProvider returns the chapter theory/exercises a variant is built from.
type SourceProvider interface {
	Source(ctx context.Context, kind practice.Kind, number int) (practice.Source, error)
}

// VariantGenerator runs variant generation in the background.
type VariantGenerator interface {
	Enqueue(accountID, variantID string, kind practice.Kind, number int)
}

type ListVariantsUseCase struct {
	variants practice.VariantRepository
}

func NewListVariantsUseCase(variants practice.VariantRepository) *ListVariantsUseCase {
	return &ListVariantsUseCase{variants: variants}
}

func (uc *ListVariantsUseCase) Execute(ctx context.Context, actor *iam.Actor, kind practice.Kind, number int) ([]practice.Variant, error) {
	return uc.variants.List(ctx, actor.AccountID().String(), kind, number)
}

type GenerateVariantUseCase struct {
	variants  practice.VariantRepository
	generator VariantGenerator
}

func NewGenerateVariantUseCase(variants practice.VariantRepository, generator VariantGenerator) *GenerateVariantUseCase {
	return &GenerateVariantUseCase{variants: variants, generator: generator}
}

func (uc *GenerateVariantUseCase) Execute(ctx context.Context, actor *iam.Actor, kind practice.Kind, number int) (practice.Variant, error) {
	// 1. Save a stub with status generating — the list shows it immediately; reload does not lose the job.
	variant := practice.Variant{
		ID:     uuid.NewString(),
		Kind:   kind,
		Number: number,
		Status: practice.StatusGenerating,
	}
	if err := variant.Validate(); err != nil {
		return practice.Variant{}, err
	}
	accountID := actor.AccountID().String()
	if err := uc.variants.Create(ctx, accountID, variant); err != nil {
		return practice.Variant{}, err
	}
	// 2. Generation itself runs in the background; status is read via state polling.
	uc.generator.Enqueue(accountID, variant.ID, kind, number)
	return variant, nil
}

type DeleteVariantUseCase struct {
	variants practice.VariantRepository
	progress practice.ProgressRepository
}

func NewDeleteVariantUseCase(variants practice.VariantRepository, progress practice.ProgressRepository) *DeleteVariantUseCase {
	return &DeleteVariantUseCase{variants: variants, progress: progress}
}

func (uc *DeleteVariantUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) error {
	accountID := actor.AccountID().String()
	// 1. Find the variant to know its chapter for progress cleanup.
	variant, err := uc.variants.Get(ctx, accountID, id)
	if err != nil {
		return err
	}
	// 2. Delete the variant and its related progress.
	if err := uc.variants.Delete(ctx, accountID, id); err != nil {
		return err
	}
	return uc.progress.Delete(ctx, accountID, variant.Kind, variant.Number, id)
}
