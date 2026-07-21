package films

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/shared"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const columns = `id, title, description, poster_path, duration_ms, status, error, kind, level, series_title, season, episode, audio_variants, subtitles, created_at`

func (s *Store) List(ctx context.Context) ([]films.Film, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+columns+` FROM films ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list films: %w", err)
	}
	defer rows.Close()
	out := []films.Film{}
	for rows.Next() {
		film, err := scanFilm(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, film)
	}
	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, id string) (films.Film, error) {
	film, err := scanFilm(s.pool.QueryRow(ctx, `SELECT `+columns+` FROM films WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return films.Film{}, shared.ErrNotFound
	}
	if err != nil {
		return films.Film{}, fmt.Errorf("get film: %w", err)
	}
	return film, nil
}

func (s *Store) Create(ctx context.Context, film films.Film) error {
	variants, subtitles, err := marshalTracks(film)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO films (id, title, description, poster_path, duration_ms, status, error, kind, level, series_title, season, episode, audio_variants, subtitles, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		film.ID, film.Title, film.Description, film.PosterPath, film.DurationMs, film.Status, film.Error, kindOrDefault(film.Kind), film.Level, film.SeriesTitle, film.Season, film.Episode, variants, subtitles, film.CreatedAt)
	if err != nil {
		return fmt.Errorf("create film: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, film films.Film) error {
	variants, subtitles, err := marshalTracks(film)
	if err != nil {
		return err
	}
	ct, err := s.pool.Exec(ctx, `UPDATE films
		SET title=$2, description=$3, poster_path=$4, duration_ms=$5, status=$6, error=$7, kind=$8, level=$9, series_title=$10, season=$11, episode=$12, audio_variants=$13, subtitles=$14
		WHERE id=$1`,
		film.ID, film.Title, film.Description, film.PosterPath, film.DurationMs, film.Status, film.Error, kindOrDefault(film.Kind), film.Level, film.SeriesTitle, film.Season, film.Episode, variants, subtitles)
	if err != nil {
		return fmt.Errorf("update film: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM films WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete film: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) ListSeries(ctx context.Context) ([]films.Series, error) {
	rows, err := s.pool.Query(ctx, `SELECT title, description, poster_path FROM film_series`)
	if err != nil {
		return nil, fmt.Errorf("list series: %w", err)
	}
	defer rows.Close()
	out := []films.Series{}
	for rows.Next() {
		var v films.Series
		if err := rows.Scan(&v.Title, &v.Description, &v.PosterPath); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) GetSeries(ctx context.Context, title string) (films.Series, error) {
	var v films.Series
	err := s.pool.QueryRow(ctx, `SELECT title, description, poster_path FROM film_series WHERE title = $1`, title).
		Scan(&v.Title, &v.Description, &v.PosterPath)
	if errors.Is(err, pgx.ErrNoRows) {
		return films.Series{}, shared.ErrNotFound
	}
	if err != nil {
		return films.Series{}, fmt.Errorf("get series: %w", err)
	}
	return v, nil
}

func (s *Store) UpsertSeries(ctx context.Context, v films.Series) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO film_series (title, description, poster_path) VALUES ($1, $2, $3)
		ON CONFLICT (title) DO UPDATE SET description = EXCLUDED.description, poster_path = EXCLUDED.poster_path`,
		v.Title, v.Description, v.PosterPath)
	if err != nil {
		return fmt.Errorf("upsert series: %w", err)
	}
	return nil
}

func (s *Store) RenameSeries(ctx context.Context, oldTitle, newTitle string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("rename series: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE films SET series_title = $2 WHERE series_title = $1`, oldTitle, newTitle); err != nil {
		return fmt.Errorf("rename series episodes: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE film_series SET title = $2 WHERE title = $1`, oldTitle, newTitle); err != nil {
		return fmt.Errorf("rename series meta: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Store) SaveProgress(ctx context.Context, ownerID, filmID string, positionMs int, updatedAt time.Time) error {
	ct, err := s.pool.Exec(ctx, `INSERT INTO film_progress (owner_id, film_id, position_ms, updated_at)
		SELECT $1, id, $3, $4 FROM films WHERE id = $2
		ON CONFLICT (owner_id, film_id) DO UPDATE SET position_ms = EXCLUDED.position_ms, updated_at = EXCLUDED.updated_at`,
		ownerID, filmID, positionMs, updatedAt)
	if err != nil {
		return fmt.Errorf("save film progress: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) ListProgress(ctx context.Context, ownerID string) (map[string]int, error) {
	rows, err := s.pool.Query(ctx, `SELECT film_id, position_ms FROM film_progress WHERE owner_id = $1`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list film progress: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var filmID string
		var positionMs int
		if err := rows.Scan(&filmID, &positionMs); err != nil {
			return nil, err
		}
		out[filmID] = positionMs
	}
	return out, rows.Err()
}

func kindOrDefault(kind string) string {
	if kind == films.KindSeries {
		return films.KindSeries
	}
	return films.KindFilm
}

func marshalTracks(film films.Film) ([]byte, []byte, error) {
	variants, err := json.Marshal(film.AudioVariants)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal audio variants: %w", err)
	}
	subtitles, err := json.Marshal(film.Subtitles)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal subtitles: %w", err)
	}
	return variants, subtitles, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanFilm(row scannable) (films.Film, error) {
	var film films.Film
	var variants, subtitles []byte
	if err := row.Scan(&film.ID, &film.Title, &film.Description, &film.PosterPath, &film.DurationMs, &film.Status, &film.Error, &film.Kind, &film.Level, &film.SeriesTitle, &film.Season, &film.Episode, &variants, &subtitles, &film.CreatedAt); err != nil {
		return films.Film{}, err
	}
	film.AudioVariants = []films.AudioVariant{}
	film.Subtitles = []films.SubtitleTrack{}
	if len(variants) > 0 {
		if err := json.Unmarshal(variants, &film.AudioVariants); err != nil {
			return films.Film{}, fmt.Errorf("unmarshal audio variants: %w", err)
		}
	}
	if len(subtitles) > 0 {
		if err := json.Unmarshal(subtitles, &film.Subtitles); err != nil {
			return films.Film{}, fmt.Errorf("unmarshal subtitles: %w", err)
		}
	}
	return film, nil
}
