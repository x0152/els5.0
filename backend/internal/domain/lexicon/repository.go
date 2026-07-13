package lexicon

import "context"

type Analyzer interface {
	Analyze(ctx context.Context, html string) (Analysis, error)
}

type Repository interface {
	SaveSentence(ctx context.Context, mediaID string, a Analysis, segments []Segment) error
	SaveSubtitle(ctx context.Context, mediaID string, cues []Cue, a Analysis) error
	DeleteByMedia(ctx context.Context, mediaID string) error
	FindOccurrences(ctx context.Context, lemmas []string) ([]LemmaOccurrences, error)
}
