package usecases

import (
	"context"

	"github.com/els/backend/internal/application/quest/runtime"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type SuggestNativeReplyUseCase struct {
	dialog *runtime.Dialog
}

func NewSuggestNativeReplyUseCase(dialog *runtime.Dialog) *SuggestNativeReplyUseCase {
	return &SuggestNativeReplyUseCase{dialog: dialog}
}

type SuggestNativeReplyResult struct {
	Variants []string
}

func (uc *SuggestNativeReplyUseCase) Execute(
	ctx context.Context,
	actor *iam.Actor,
	missionID string,
	text string,
) (SuggestNativeReplyResult, error) {
	// 1. Allow only authenticated player requests.
	if actor == nil {
		return SuggestNativeReplyResult{}, shared.ErrUnauthorized
	}

	// 2. Generate native-like variants in mission context.
	variants, err := uc.dialog.NativeReplyVariants(ctx, actor.AccountID().String(), missionID, text)
	if err != nil {
		return SuggestNativeReplyResult{}, err
	}

	return SuggestNativeReplyResult{Variants: variants}, nil
}
