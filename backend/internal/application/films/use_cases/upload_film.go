package usecases

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/lexicon"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared/ports"
)

type UploadFilmUseCase struct {
	films      films.Repository
	storage    media.Storage
	transcoder ports.Transcoder
	analyzer   lexicon.Analyzer
	lex        lexicon.Repository
	bucket     string
	tempDir    string
	logger     *slog.Logger
}

func NewUploadFilmUseCase(repo films.Repository, storage media.Storage, transcoder ports.Transcoder, analyzer lexicon.Analyzer, lex lexicon.Repository, bucket, tempDir string, logger *slog.Logger) *UploadFilmUseCase {
	return &UploadFilmUseCase{films: repo, storage: storage, transcoder: transcoder, analyzer: analyzer, lex: lex, bucket: bucket, tempDir: tempDir, logger: logger}
}

type UploadAsset struct {
	Data        []byte
	ContentType string
	Filename    string
}

type UploadFilmCommand struct {
	Title         string
	Filename      string
	VideoTempPath string
	Kind          string
	Level         string
	SeriesTitle   string
	Season        int
	Episode       int
	Poster        *UploadAsset
	SubtitleSRT   []byte
}

func (uc *UploadFilmUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd UploadFilmCommand) (films.Film, error) {
	// 1. Only a global admin uploads films.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return films.Film{}, err
	}

	// 2. Immediately save the film as processing; tracks will arrive later.
	kind := films.KindFilm
	if cmd.Kind == films.KindSeries {
		kind = films.KindSeries
	}
	level, err := films.ParseLevel(cmd.Level)
	if err != nil {
		_ = os.Remove(cmd.VideoTempPath)
		return films.Film{}, err
	}
	film := films.Film{
		ID:            uuid.NewString(),
		Title:         filmTitle(cmd, kind),
		Status:        films.StatusProcessing,
		Kind:          kind,
		Level:         level,
		SeriesTitle:   strings.TrimSpace(cmd.SeriesTitle),
		Season:        cmd.Season,
		Episode:       cmd.Episode,
		AudioVariants: []films.AudioVariant{},
		Subtitles:     []films.SubtitleTrack{},
		CreatedAt:     time.Now().UTC(),
	}
	if err := film.Validate(); err != nil {
		_ = os.Remove(cmd.VideoTempPath)
		return films.Film{}, err
	}
	if err := uc.films.Create(ctx, film); err != nil {
		_ = os.Remove(cmd.VideoTempPath)
		return films.Film{}, err
	}

	// 3. Transcode in the background; the client polls for status.
	go uc.process(context.WithoutCancel(ctx), film, cmd)
	return film, nil
}

func (uc *UploadFilmUseCase) process(ctx context.Context, film films.Film, cmd UploadFilmCommand) {
	defer os.Remove(cmd.VideoTempPath)

	// 1. On error, mark the film failed and save the reason.
	if err := uc.transcode(ctx, &film, cmd); err != nil {
		uc.logger.Error("films transcode failed", slog.String("film", film.ID), slog.String("err", err.Error()))
		film.Status = films.StatusFailed
		film.Error = err.Error()
		if uErr := uc.films.Update(ctx, film); uErr != nil {
			uc.logger.Error("films mark failed", slog.String("film", film.ID), slog.String("err", uErr.Error()))
		}
		return
	}

	// 2. Success: tracks are ready, move to ready.
	film.Status = films.StatusReady
	if err := uc.films.Update(ctx, film); err != nil {
		uc.logger.Error("films mark ready", slog.String("film", film.ID), slog.String("err", err.Error()))
	}
}

func (uc *UploadFilmUseCase) transcode(ctx context.Context, film *films.Film, cmd UploadFilmCommand) error {
	// 1. Read source tracks via ffprobe.
	probe, err := uc.transcoder.Probe(ctx, cmd.VideoTempPath)
	if err != nil {
		return err
	}
	if len(probe.Audio) == 0 {
		return fmt.Errorf("no audio tracks found")
	}

	// 2. For each audio track make a separate mp4; copy video if it is already h264.
	copyVideo := probe.VideoCodec == "h264"
	for i, a := range probe.Audio {
		out := filepath.Join(uc.tempDir, fmt.Sprintf("%s-audio-%d.mp4", film.ID, i))
		if err := uc.transcoder.BuildAudioVariant(ctx, cmd.VideoTempPath, a.Index, copyVideo, out); err != nil {
			return err
		}
		path, err := uc.putFile(ctx, fmt.Sprintf("%s/audio-%d.mp4", film.ID, i), out, "video/mp4")
		os.Remove(out)
		if err != nil {
			return err
		}
		film.AudioVariants = append(film.AudioVariants, films.AudioVariant{
			Lang:  a.Lang,
			Label: trackLabel(a, "Audio", i),
			Path:  path,
		})
	}

	// 3. Extract text subtitles to srt and parse into cues.
	for i, s := range probe.Subs {
		data, err := uc.transcoder.ExtractSubtitleSRT(ctx, cmd.VideoTempPath, s.Index)
		if err != nil {
			uc.logger.Warn("films subtitle extract failed", slog.String("film", film.ID), slog.String("err", err.Error()))
			continue
		}
		cues := films.ParseSRT(data)
		if len(cues) == 0 {
			continue
		}
		film.Subtitles = append(film.Subtitles, films.SubtitleTrack{
			Lang:  s.Lang,
			Label: trackLabel(s, "Subtitles", i),
			Cues:  cues,
		})
	}

	// 4. A manually uploaded srt becomes the first track.
	if len(cmd.SubtitleSRT) > 0 {
		if cues := films.ParseSRT(cmd.SubtitleSRT); len(cues) > 0 {
			film.Subtitles = append([]films.SubtitleTrack{{Lang: "en", Label: "Uploaded", Cues: cues}}, film.Subtitles...)
		}
	}

	// 5. Take duration from ffprobe, otherwise from the last subtitle cue.
	film.DurationMs = probe.DurationMs
	if film.DurationMs == 0 {
		for _, t := range film.Subtitles {
			if n := len(t.Cues); n > 0 && t.Cues[n-1].EndMs > film.DurationMs {
				film.DurationMs = t.Cues[n-1].EndMs
			}
		}
	}

	// 6. Poster: uploaded one, otherwise a frame from the video.
	if cmd.Poster != nil && len(cmd.Poster.Data) > 0 {
		path, err := uc.putBytes(ctx, fmt.Sprintf("%s/poster%s", film.ID, ext(cmd.Poster.Filename, ".jpg")), cmd.Poster.Data, cmd.Poster.ContentType)
		if err != nil {
			return err
		}
		film.PosterPath = path
	} else {
		thumb := filepath.Join(uc.tempDir, film.ID+"-thumb.jpg")
		seek := float64(film.DurationMs) / 1000 * (0.4 + rand.Float64()*0.2) // #nosec G404 -- random thumbnail timestamp, not security sensitive.
		if err := uc.transcoder.ExtractThumbnail(ctx, cmd.VideoTempPath, thumb, seek); err == nil {
			if path, err := uc.putFile(ctx, film.ID+"/poster.jpg", thumb, "image/jpeg"); err == nil {
				film.PosterPath = path
			}
			os.Remove(thumb)
		}
	}

	// 7. The English subtitle track goes to spaCy for lexeme parsing.
	return uc.analyze(ctx, film.ID, film.Subtitles)
}

func (uc *UploadFilmUseCase) analyze(ctx context.Context, mediaID string, tracks []films.SubtitleTrack) error {
	if uc.analyzer == nil || uc.lex == nil {
		return nil
	}
	track, ok := films.PickEnglishSubtitle(tracks)
	if !ok || len(track.Cues) == 0 {
		return nil
	}
	cues := make([]lexicon.Cue, 0, len(track.Cues))
	for _, c := range track.Cues {
		cues = append(cues, lexicon.Cue{Index: c.Index, StartMs: c.StartMs, EndMs: c.EndMs, Text: c.Text})
	}
	input, segs := lexicon.BuildSubtitleInput(cues)
	raw, err := uc.analyzer.Analyze(ctx, input)
	if err != nil {
		return fmt.Errorf("lexicon analyze: %w", err)
	}
	analysis := lexicon.AttachSubtitleSegments(lexicon.MapAnalysis(raw), segs)
	if err := uc.lex.SaveSubtitle(ctx, mediaID, cues, analysis); err != nil {
		return fmt.Errorf("lexicon save: %w", err)
	}
	return nil
}

func (uc *UploadFilmUseCase) putFile(ctx context.Context, key, localPath, contentType string) (string, error) {
	// #nosec G304 -- localPath is a temp transcoder output generated by this process.
	f, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	return uc.put(ctx, key, f, stat.Size(), contentType)
}

func (uc *UploadFilmUseCase) putBytes(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	return uc.put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
}

func (uc *UploadFilmUseCase) put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	path, err := media.NewPath(uc.bucket + "/" + key)
	if err != nil {
		return "", err
	}
	if err := uc.storage.Put(ctx, path, r, media.PutOptions{ContentType: contentType, Size: size}); err != nil {
		return "", err
	}
	return path.String(), nil
}

func trackLabel(s ports.MediaStream, fallback string, i int) string {
	if t := strings.TrimSpace(s.Title); t != "" {
		return t
	}
	if l := strings.TrimSpace(s.Lang); l != "" {
		return strings.ToUpper(l)
	}
	return fmt.Sprintf("%s %d", fallback, i+1)
}

func titleOrFilename(title, filename string) string {
	if t := strings.TrimSpace(title); t != "" {
		return t
	}
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func filmTitle(cmd UploadFilmCommand, kind string) string {
	if t := strings.TrimSpace(cmd.Title); t != "" {
		return t
	}
	if kind == films.KindSeries {
		return fmt.Sprintf("S%dE%d", cmd.Season, cmd.Episode)
	}
	return titleOrFilename(cmd.Title, cmd.Filename)
}

func ext(filename, fallback string) string {
	if e := filepath.Ext(filename); e != "" {
		return strings.ToLower(e)
	}
	return fallback
}
