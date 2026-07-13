package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/lexicon"
	"github.com/els/backend/internal/domain/shared"
)

type OccurrenceResult struct {
	Common     bool
	Total      int
	MediaCount int
	Media      []lexicon.MediaOccurrence
}

type OccurrencesUseCase struct {
	lexicon LexiconReader
}

func NewOccurrencesUseCase(lexicon LexiconReader) *OccurrencesUseCase {
	return &OccurrencesUseCase{lexicon: lexicon}
}

func (uc *OccurrencesUseCase) Execute(ctx context.Context, actor *iam.Actor, text string) (OccurrenceResult, error) {
	if actor == nil {
		return OccurrenceResult{}, shared.ErrUnauthorized
	}
	lemma := strings.ToLower(strings.TrimSpace(text))
	if lemma == "" {
		return OccurrenceResult{}, shared.Validation(fmt.Errorf("text: must not be empty"))
	}
	if uc.lexicon == nil {
		return OccurrenceResult{}, nil
	}

	found, err := uc.lexicon.FindOccurrences(ctx, []string{lemma})
	if err != nil {
		return OccurrenceResult{}, err
	}

	media, total, mediaCount, stop := combineMedia(found)
	res := OccurrenceResult{
		Total:      total,
		MediaCount: mediaCount,
		Media:      media,
		Common:     stop || mediaCount > commonMediaThreshold,
	}
	if res.Common {
		res.Media = nil
	}
	return res, nil
}
