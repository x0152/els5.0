package usecases

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
)

type UpdateFilmUseCase struct {
	films   films.Repository
	storage media.Storage
	bucket  string
}

func NewUpdateFilmUseCase(repo films.Repository, storage media.Storage, bucket string) *UpdateFilmUseCase {
	return &UpdateFilmUseCase{films: repo, storage: storage, bucket: bucket}
}

type UpdateFilmCommand struct {
	Title       string
	Description string
	Level       string
	Poster      *UploadAsset
}

func (uc *UpdateFilmUseCase) Execute(ctx context.Context, actor *iam.Actor, id string, cmd UpdateFilmCommand) (films.Film, error) {
	// 1. Only a global admin edits films.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return films.Film{}, err
	}

	// 2. Read the film to preserve the other fields.
	film, err := uc.films.Get(ctx, id)
	if err != nil {
		return films.Film{}, err
	}

	// 3. Update the title, description and level.
	film.Title = strings.TrimSpace(cmd.Title)
	film.Description = strings.TrimSpace(cmd.Description)
	level, err := films.ParseLevel(cmd.Level)
	if err != nil {
		return films.Film{}, err
	}
	film.Level = level
	if err := film.Validate(); err != nil {
		return films.Film{}, err
	}

	// 4. A new poster is optional: re-upload and clean up the old one if the path changed.
	if cmd.Poster != nil && len(cmd.Poster.Data) > 0 {
		key := fmt.Sprintf("%s/poster-%s%s", film.ID, uuid.NewString(), ext(cmd.Poster.Filename, ".jpg"))
		path, err := media.NewPath(uc.bucket + "/" + key)
		if err != nil {
			return films.Film{}, err
		}
		if err := uc.storage.Put(ctx, path, bytes.NewReader(cmd.Poster.Data), media.PutOptions{ContentType: cmd.Poster.ContentType, Size: int64(len(cmd.Poster.Data))}); err != nil {
			return films.Film{}, err
		}
		if film.PosterPath != "" && film.PosterPath != path.String() {
			if old, err := media.NewPath(film.PosterPath); err == nil {
				_ = uc.storage.Delete(ctx, old)
			}
		}
		film.PosterPath = path.String()
	}

	// 5. Persist the changes.
	if err := uc.films.Update(ctx, film); err != nil {
		return films.Film{}, err
	}
	return film, nil
}
