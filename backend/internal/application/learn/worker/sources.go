package worker

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/practice"
	bookrepo "github.com/els/backend/internal/infrastructure/repositories/book"
)

// Sources reads chapter content from the book store so variants can be
// generated and free answers checked against the original theory.
type Sources struct {
	chapters *bookrepo.Store
}

func NewSources(pool *pgxpool.Pool) *Sources {
	return &Sources{chapters: bookrepo.NewStore(pool)}
}

func (s *Sources) Source(ctx context.Context, kind practice.Kind, number int) (practice.Source, error) {
	c, err := s.chapters.GetByNumber(ctx, string(kind), number)
	if err != nil {
		return practice.Source{}, err
	}
	return practice.Source{Title: c.Title, Theory: c.Theory, Exercises: c.Exercises}, nil
}
