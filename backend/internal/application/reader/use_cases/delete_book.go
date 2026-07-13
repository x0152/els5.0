package usecases

import (
	"context"
	"log/slog"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/lexicon"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type DeleteBookUseCase struct {
	books   reader.Repository
	storage media.Storage
	lex     lexicon.Repository
	logger  *slog.Logger
}

func NewDeleteBookUseCase(repo reader.Repository, storage media.Storage, lex lexicon.Repository, logger *slog.Logger) *DeleteBookUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &DeleteBookUseCase{books: repo, storage: storage, lex: lex, logger: logger}
}

func (uc *DeleteBookUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) error {
	// 1. The library is shared: any authenticated user may delete a book.
	if actor == nil {
		return shared.ErrUnauthorized
	}
	book, err := uc.books.Get(ctx, actor.AccountID().String(), id)
	if err != nil {
		return err
	}

	// 2. Clean up files and lexicon best-effort: the book is deleted either way.
	if uc.storage != nil {
		for _, raw := range []string{book.ContentPath, book.CoverPath} {
			if raw == "" {
				continue
			}
			path, err := media.NewPath(raw)
			if err != nil {
				continue
			}
			if err := uc.storage.Delete(ctx, path); err != nil {
				uc.logger.Warn("reader: delete book file failed", slog.String("path", raw), slog.String("err", err.Error()))
			}
		}
	}
	if uc.lex != nil {
		if err := uc.lex.DeleteByMedia(ctx, id); err != nil {
			uc.logger.Warn("reader: delete book lexicon failed", slog.String("book", id), slog.String("err", err.Error()))
		}
	}

	// 3. Delete the book record (all readers' positions cascade away).
	return uc.books.Delete(ctx, id)
}
