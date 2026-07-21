package api

import (
	"github.com/danielgtaylor/huma/v2"

	authx "github.com/els/backend/internal/utils/auth"
)

type CueOutput struct {
	Index   int    `json:"index"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
	Text    string `json:"text"`
}

type SubtitleTrackOutput struct {
	Lang  string      `json:"lang"`
	Label string      `json:"label"`
	Cues  []CueOutput `json:"cues"`
}

type AudioTrackOutput struct {
	Lang  string `json:"lang"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

type FilmSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	PosterURL   string `json:"poster_url,omitempty"`
	DurationMs  int    `json:"duration_ms"`
	PositionMs  int    `json:"position_ms"`
	Status      string `json:"status"`
	Kind        string `json:"kind"`
	Level       string `json:"level"`
	SeriesTitle string `json:"series_title,omitempty"`
	Season      int    `json:"season"`
	Episode     int    `json:"episode"`
	CreatedAt   string `json:"created_at"`
}

type FilmOutput struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description,omitempty"`
	PosterURL   string                `json:"poster_url,omitempty"`
	DurationMs  int                   `json:"duration_ms"`
	PositionMs  int                   `json:"position_ms"`
	Status      string                `json:"status"`
	Error       string                `json:"error,omitempty"`
	Kind        string                `json:"kind"`
	Level       string                `json:"level"`
	SeriesTitle string                `json:"series_title,omitempty"`
	Season      int                   `json:"season"`
	Episode     int                   `json:"episode"`
	AudioTracks []AudioTrackOutput    `json:"audio_tracks"`
	Subtitles   []SubtitleTrackOutput `json:"subtitles"`
	CreatedAt   string                `json:"created_at"`
}

type FilmsOutput struct {
	Items []FilmSummary `json:"items"`
}

type SeriesOutput struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	PosterURL   string `json:"poster_url,omitempty"`
}

type SeriesListOutput struct {
	Items []SeriesOutput `json:"items"`
}

type ListSeriesInput struct {
	authx.BearerInput
}

type UpdateSeriesForm struct {
	Poster huma.FormFile `form:"poster" contentType:"image/png,image/jpeg,image/webp" required:"false"`
}

type UpdateSeriesInput struct {
	authx.BearerInput
	RawBody huma.MultipartFormFiles[UpdateSeriesForm]
}

type ListFilmsInput struct {
	authx.BearerInput
}

type GetFilmInput struct {
	authx.BearerInput
	ID string `path:"id"`
}

type UploadFilmForm struct {
	Video     huma.FormFile `form:"video" required:"true"`
	Subtitles huma.FormFile `form:"subtitles" required:"false"`
	Poster    huma.FormFile `form:"poster" contentType:"image/png,image/jpeg,image/webp" required:"false"`
}

type UploadFilmInput struct {
	authx.BearerInput
	RawBody huma.MultipartFormFiles[UploadFilmForm]
}

type UpdateFilmForm struct {
	Poster huma.FormFile `form:"poster" contentType:"image/png,image/jpeg,image/webp" required:"false"`
}

type UpdateFilmInput struct {
	authx.BearerInput
	ID      string `path:"id"`
	RawBody huma.MultipartFormFiles[UpdateFilmForm]
}

type DeleteFilmInput struct {
	authx.BearerInput
	ID string `path:"id"`
}

type SaveProgressInput struct {
	authx.BearerInput
	ID   string `path:"id"`
	Body struct {
		PositionMs int `json:"position_ms" minimum:"0"`
	}
}

type SaveProgressOutput struct {
	OK bool `json:"ok"`
}

type DeleteFilmOutput struct {
	OK bool `json:"ok"`
}
