package api

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/films/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator
	ListFilms     *usecases.ListFilmsUseCase
	GetFilm       *usecases.GetFilmUseCase
	UploadFilm    *usecases.UploadFilmUseCase
	UpdateFilm    *usecases.UpdateFilmUseCase
	DeleteFilm    *usecases.DeleteFilmUseCase
	SaveProgress  *usecases.SaveProgressUseCase
	ListSeries    *usecases.ListSeriesUseCase
	UpdateSeries  *usecases.UpdateSeriesUseCase
	MediaURLs     media.PublicURL
	TempDir       string
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listFilms",
		Method:      http.MethodGet,
		Path:        "/api/v1/films",
		Summary:     "List films",
		Tags:        []string{"films"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListFilmsInput) (FilmsOutput, error) {
		list, progress, err := deps.ListFilms.Execute(ctx, actor)
		if err != nil {
			return FilmsOutput{}, err
		}
		return toFilmsOutput(list, progress, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getFilm",
		Method:      http.MethodGet,
		Path:        "/api/v1/films/{id}",
		Summary:     "Get a film with audio and subtitle tracks",
		Tags:        []string{"films"},
	}, func(ctx context.Context, actor *iam.Actor, in *GetFilmInput) (FilmOutput, error) {
		film, positionMs, err := deps.GetFilm.Execute(ctx, actor, in.ID)
		if err != nil {
			return FilmOutput{}, err
		}
		return toFilmOutput(film, positionMs, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "saveFilmProgress",
		Method:      http.MethodPut,
		Path:        "/api/v1/films/{id}/progress",
		Summary:     "Save the watch position for a film",
		Tags:        []string{"films"},
	}, func(ctx context.Context, actor *iam.Actor, in *SaveProgressInput) (SaveProgressOutput, error) {
		if err := deps.SaveProgress.Execute(ctx, actor, in.ID, in.Body.PositionMs); err != nil {
			return SaveProgressOutput{}, err
		}
		return SaveProgressOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "uploadFilm",
		Method:        http.MethodPost,
		Path:          "/api/v1/films",
		Summary:       "Upload a film; tracks are transcoded in the background (global admin only)",
		Tags:          []string{"films"},
		DefaultStatus: http.StatusAccepted,
	}, func(ctx context.Context, actor *iam.Actor, in *UploadFilmInput) (FilmOutput, error) {
		if deps.UploadFilm == nil {
			return FilmOutput{}, huma.Error503ServiceUnavailable("film upload is not configured")
		}
		form := in.RawBody.Data()
		if form == nil || !form.Video.IsSet {
			return FilmOutput{}, huma.Error400BadRequest("video file is required")
		}
		defer form.Video.Close()

		tempPath, err := saveTemp(deps.TempDir, form.Video, form.Video.Filename)
		if err != nil {
			return FilmOutput{}, huma.Error500InternalServerError("failed to buffer upload")
		}

		cmd := usecases.UploadFilmCommand{
			Title:         formValue(in.RawBody.Form, "title"),
			Filename:      form.Video.Filename,
			VideoTempPath: tempPath,
			Kind:          formValue(in.RawBody.Form, "kind"),
			SeriesTitle:   formValue(in.RawBody.Form, "series_title"),
			Season:        formInt(in.RawBody.Form, "season"),
			Episode:       formInt(in.RawBody.Form, "episode"),
		}
		if form.Subtitles.IsSet {
			defer form.Subtitles.Close()
			if data, err := io.ReadAll(io.LimitReader(form.Subtitles, 10<<20)); err == nil {
				cmd.SubtitleSRT = data
			}
		}
		if form.Poster.IsSet {
			defer form.Poster.Close()
			if data, err := io.ReadAll(io.LimitReader(form.Poster, 20<<20)); err == nil {
				cmd.Poster = &usecases.UploadAsset{Data: data, ContentType: form.Poster.ContentType, Filename: form.Poster.Filename}
			}
		}

		film, err := deps.UploadFilm.Execute(ctx, actor, cmd)
		if err != nil {
			os.Remove(tempPath)
			return FilmOutput{}, err
		}
		return toFilmOutput(film, 0, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "updateFilm",
		Method:      http.MethodPatch,
		Path:        "/api/v1/films/{id}",
		Summary:     "Update a film's title, description and poster (global admin only)",
		Tags:        []string{"films"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateFilmInput) (FilmOutput, error) {
		if deps.UpdateFilm == nil {
			return FilmOutput{}, huma.Error503ServiceUnavailable("film editing is not configured")
		}
		form := in.RawBody.Data()
		cmd := usecases.UpdateFilmCommand{
			Title:       formValue(in.RawBody.Form, "title"),
			Description: formValue(in.RawBody.Form, "description"),
		}
		if form != nil && form.Poster.IsSet {
			defer form.Poster.Close()
			if data, err := io.ReadAll(io.LimitReader(form.Poster, 20<<20)); err == nil {
				cmd.Poster = &usecases.UploadAsset{Data: data, ContentType: form.Poster.ContentType, Filename: form.Poster.Filename}
			}
		}
		film, err := deps.UpdateFilm.Execute(ctx, actor, in.ID, cmd)
		if err != nil {
			return FilmOutput{}, err
		}
		return toFilmOutput(film, 0, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listSeries",
		Method:      http.MethodGet,
		Path:        "/api/v1/series",
		Summary:     "List series metadata (cover and description)",
		Tags:        []string{"films"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListSeriesInput) (SeriesListOutput, error) {
		list, err := deps.ListSeries.Execute(ctx, actor)
		if err != nil {
			return SeriesListOutput{}, err
		}
		return toSeriesListOutput(list, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "updateSeries",
		Method:      http.MethodPatch,
		Path:        "/api/v1/series",
		Summary:     "Update a series title, description and cover (global admin only)",
		Tags:        []string{"films"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateSeriesInput) (SeriesOutput, error) {
		if deps.UpdateSeries == nil {
			return SeriesOutput{}, huma.Error503ServiceUnavailable("series editing is not configured")
		}
		form := in.RawBody.Data()
		cmd := usecases.UpdateSeriesCommand{
			Title:       formValue(in.RawBody.Form, "title"),
			NewTitle:    formValue(in.RawBody.Form, "new_title"),
			Description: formValue(in.RawBody.Form, "description"),
		}
		if form != nil && form.Poster.IsSet {
			defer form.Poster.Close()
			if data, err := io.ReadAll(io.LimitReader(form.Poster, 20<<20)); err == nil {
				cmd.Poster = &usecases.UploadAsset{Data: data, ContentType: form.Poster.ContentType, Filename: form.Poster.Filename}
			}
		}
		series, err := deps.UpdateSeries.Execute(ctx, actor, cmd)
		if err != nil {
			return SeriesOutput{}, err
		}
		return toSeriesOutput(series, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "deleteFilm",
		Method:      http.MethodDelete,
		Path:        "/api/v1/films/{id}",
		Summary:     "Delete a film (global admin only)",
		Tags:        []string{"films"},
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteFilmInput) (DeleteFilmOutput, error) {
		if err := deps.DeleteFilm.Execute(ctx, actor, in.ID); err != nil {
			return DeleteFilmOutput{}, err
		}
		return DeleteFilmOutput{OK: true}, nil
	})
}

func saveTemp(dir string, r io.Reader, filename string) (string, error) {
	f, err := os.CreateTemp(dir, "film-upload-*"+filepath.Ext(filename))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func formValue(form *multipart.Form, key string) string {
	if form == nil {
		return ""
	}
	if vals := form.Value[key]; len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func formInt(form *multipart.Form, key string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(formValue(form, key)))
	return n
}
