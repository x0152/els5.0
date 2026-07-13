package usecases

import (
	"context"
	"log/slog"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/lexicon"
	"github.com/els/backend/internal/domain/media"
)

type DeleteFilmUseCase struct {
	films   films.Repository
	storage media.Storage
	lex     lexicon.Repository
	logger  *slog.Logger
}

func NewDeleteFilmUseCase(repo films.Repository, storage media.Storage, lex lexicon.Repository, logger *slog.Logger) *DeleteFilmUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &DeleteFilmUseCase{films: repo, storage: storage, lex: lex, logger: logger}
}

func (uc *DeleteFilmUseCase) Execute(ctx context.Context, actor *iam.Actor, id string) error {
	// 1. Only a global admin deletes films.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return err
	}

	// 2. First read the film to know the storage paths.
	film, err := uc.films.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.films.Delete(ctx, id); err != nil {
		return err
	}
	if uc.lex != nil {
		if err := uc.lex.DeleteByMedia(ctx, id); err != nil {
			uc.logger.Warn("films: delete film lexicon failed", slog.String("film", id), slog.String("err", err.Error()))
		}
	}

	// 3. Clean up the poster and all audio variants from storage.
	paths := []string{film.PosterPath}
	for _, v := range film.AudioVariants {
		paths = append(paths, v.Path)
	}
	for _, raw := range paths {
		if raw == "" {
			continue
		}
		path, err := media.NewPath(raw)
		if err != nil {
			continue
		}
		if err := uc.storage.Delete(ctx, path); err != nil {
			uc.logger.Warn("films: delete film file failed", slog.String("path", raw), slog.String("err", err.Error()))
		}
	}
	return nil
}
