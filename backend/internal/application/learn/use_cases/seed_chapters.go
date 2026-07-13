package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/book"
)

type SeedChaptersUseCase struct {
	chapters book.Repository
	books    []book.Book
	seed     []book.Chapter
}

func NewSeedChaptersUseCase(chapters book.Repository, books []book.Book, seed []book.Chapter) *SeedChaptersUseCase {
	return &SeedChaptersUseCase{chapters: chapters, books: books, seed: seed}
}

func (uc *SeedChaptersUseCase) Execute(ctx context.Context) error {
	// 1. Ensure books (DB entities) exist.
	for _, b := range uc.books {
		if err := uc.chapters.EnsureBook(ctx, b); err != nil {
			return err
		}
	}

	// 2. Seed starter chapters only into an empty-per-book database.
	seeded := map[string]bool{}
	for _, chapter := range uc.seed {
		if _, ok := seeded[chapter.Book]; !ok {
			count, err := uc.chapters.Count(ctx, chapter.Book)
			if err != nil {
				return err
			}
			seeded[chapter.Book] = count > 0
		}
		if seeded[chapter.Book] {
			continue
		}
		chapter.ID = uuid.NewString()
		if err := uc.chapters.Create(ctx, chapter); err != nil {
			return err
		}
	}
	return nil
}
