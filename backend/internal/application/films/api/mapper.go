package api

import (
	"time"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/media"
)

func toFilmSummary(film films.Film, positionMs int, urls media.PublicURL) FilmSummary {
	return FilmSummary{
		ID:          film.ID,
		Title:       film.Title,
		Description: film.Description,
		PosterURL:   buildURL(urls, film.PosterPath),
		DurationMs:  film.DurationMs,
		PositionMs:  positionMs,
		Status:      film.Status,
		Kind:        kindOrDefault(film.Kind),
		SeriesTitle: film.SeriesTitle,
		Season:      film.Season,
		Episode:     film.Episode,
		CreatedAt:   film.CreatedAt.Format(time.RFC3339),
	}
}

func kindOrDefault(kind string) string {
	if kind == films.KindSeries {
		return films.KindSeries
	}
	return films.KindFilm
}

func toFilmsOutput(list []films.Film, progress map[string]int, urls media.PublicURL) FilmsOutput {
	items := make([]FilmSummary, 0, len(list))
	for _, film := range list {
		items = append(items, toFilmSummary(film, progress[film.ID], urls))
	}
	return FilmsOutput{Items: items}
}

func toFilmOutput(film films.Film, positionMs int, urls media.PublicURL) FilmOutput {
	audio := make([]AudioTrackOutput, 0, len(film.AudioVariants))
	for _, v := range film.AudioVariants {
		audio = append(audio, AudioTrackOutput{Lang: v.Lang, Label: v.Label, URL: buildURL(urls, v.Path)})
	}
	subs := make([]SubtitleTrackOutput, 0, len(film.Subtitles))
	for _, t := range film.Subtitles {
		cues := make([]CueOutput, 0, len(t.Cues))
		for _, c := range t.Cues {
			cues = append(cues, CueOutput{Index: c.Index, StartMs: c.StartMs, EndMs: c.EndMs, Text: c.Text})
		}
		subs = append(subs, SubtitleTrackOutput{Lang: t.Lang, Label: t.Label, Cues: cues})
	}
	return FilmOutput{
		ID:          film.ID,
		Title:       film.Title,
		Description: film.Description,
		PosterURL:   buildURL(urls, film.PosterPath),
		DurationMs:  film.DurationMs,
		PositionMs:  positionMs,
		Status:      film.Status,
		Error:       film.Error,
		Kind:        kindOrDefault(film.Kind),
		SeriesTitle: film.SeriesTitle,
		Season:      film.Season,
		Episode:     film.Episode,
		AudioTracks: audio,
		Subtitles:   subs,
		CreatedAt:   film.CreatedAt.Format(time.RFC3339),
	}
}

func toSeriesOutput(s films.Series, urls media.PublicURL) SeriesOutput {
	return SeriesOutput{Title: s.Title, Description: s.Description, PosterURL: buildURL(urls, s.PosterPath)}
}

func toSeriesListOutput(list []films.Series, urls media.PublicURL) SeriesListOutput {
	items := make([]SeriesOutput, 0, len(list))
	for _, s := range list {
		items = append(items, toSeriesOutput(s, urls))
	}
	return SeriesListOutput{Items: items}
}

func buildURL(urls media.PublicURL, raw string) string {
	if raw == "" {
		return ""
	}
	return urls.BuildFromRaw(raw)
}
