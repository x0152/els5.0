package usecases

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared"
)

type ListSeriesUseCase struct {
	films films.Repository
}

func NewListSeriesUseCase(repo films.Repository) *ListSeriesUseCase {
	return &ListSeriesUseCase{films: repo}
}

func (uc *ListSeriesUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]films.Series, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	return uc.films.ListSeries(ctx)
}

type UpdateSeriesUseCase struct {
	films   films.Repository
	storage media.Storage
	bucket  string
}

func NewUpdateSeriesUseCase(repo films.Repository, storage media.Storage, bucket string) *UpdateSeriesUseCase {
	return &UpdateSeriesUseCase{films: repo, storage: storage, bucket: bucket}
}

type UpdateSeriesCommand struct {
	Title       string
	NewTitle    string
	Description string
	Poster      *UploadAsset
}

func (uc *UpdateSeriesUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd UpdateSeriesCommand) (films.Series, error) {
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return films.Series{}, err
	}

	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return films.Series{}, fmt.Errorf("%w: series.title: must not be empty", shared.ErrValidation)
	}
	newTitle := strings.TrimSpace(cmd.NewTitle)
	if newTitle == "" {
		newTitle = title
	}

	existing, err := uc.films.GetSeries(ctx, title)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return films.Series{}, err
	}

	if newTitle != title {
		if err := uc.films.RenameSeries(ctx, title, newTitle); err != nil {
			return films.Series{}, err
		}
	}

	posterPath := existing.PosterPath
	if cmd.Poster != nil && len(cmd.Poster.Data) > 0 {
		path, err := media.NewPath(uc.bucket + "/series/" + uuid.NewString() + ext(cmd.Poster.Filename, ".jpg"))
		if err != nil {
			return films.Series{}, err
		}
		if err := uc.storage.Put(ctx, path, bytes.NewReader(cmd.Poster.Data), media.PutOptions{ContentType: cmd.Poster.ContentType, Size: int64(len(cmd.Poster.Data))}); err != nil {
			return films.Series{}, err
		}
		if posterPath != "" {
			if old, err := media.NewPath(posterPath); err == nil {
				_ = uc.storage.Delete(ctx, old)
			}
		}
		posterPath = path.String()
	}

	series := films.Series{Title: newTitle, Description: strings.TrimSpace(cmd.Description), PosterPath: posterPath}
	if err := uc.films.UpsertSeries(ctx, series); err != nil {
		return films.Series{}, err
	}
	return series, nil
}
