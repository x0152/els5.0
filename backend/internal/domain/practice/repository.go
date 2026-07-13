package practice

import "context"

type VariantRepository interface {
	List(ctx context.Context, accountID string, kind Kind, number int) ([]Variant, error)
	Get(ctx context.Context, accountID, id string) (Variant, error)
	Create(ctx context.Context, accountID string, v Variant) error
	Update(ctx context.Context, accountID string, v Variant) error
	Delete(ctx context.Context, accountID, id string) error
}

type ProgressRepository interface {
	Get(ctx context.Context, accountID string, kind Kind, number int, variantKey string) (Progress, error)
	Save(ctx context.Context, accountID string, kind Kind, number int, variantKey string, p Progress) error
	Delete(ctx context.Context, accountID string, kind Kind, number int, variantKey string) error
}
