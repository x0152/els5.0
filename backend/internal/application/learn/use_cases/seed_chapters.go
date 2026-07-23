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
	// 1. Seed only once — when the books table is empty (first deploy / fresh DB).
	existing, err := uc.chapters.ListBooks(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	// 2. Insert books from embed.
	for _, b := range uc.books {
		if err := uc.chapters.EnsureBook(ctx, b); err != nil {
			return err
		}
	}

	// 3. Insert starter chapters.
	for _, chapter := range uc.seed {
		chapter.ID = uuid.NewString()
		if err := uc.chapters.Create(ctx, chapter); err != nil {
			return err
		}
	}
	return nil
}
