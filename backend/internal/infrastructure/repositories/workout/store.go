package workout

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/infrastructure/postgres"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) GetPlan(ctx context.Context, filmID string) (workout.FilmPlan, error) {
	var plan workout.FilmPlan
	var segments []byte
	err := s.pool.QueryRow(ctx, `SELECT film_id, status, error, segments, created_at FROM workout_plans WHERE film_id = $1`, filmID).
		Scan(&plan.FilmID, &plan.Status, &plan.Error, &segments, &plan.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return workout.FilmPlan{}, shared.ErrNotFound
	}
	if err != nil {
		return workout.FilmPlan{}, fmt.Errorf("get plan: %w", err)
	}
	if err := json.Unmarshal(segments, &plan.Segments); err != nil {
		return workout.FilmPlan{}, fmt.Errorf("unmarshal segments: %w", err)
	}
	return plan, nil
}

func (s *Store) SavePlan(ctx context.Context, plan workout.FilmPlan) error {
	segments, err := json.Marshal(plan.Segments)
	if err != nil {
		return fmt.Errorf("marshal segments: %w", err)
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO workout_plans (film_id, status, error, segments, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (film_id) DO UPDATE SET status = EXCLUDED.status, error = EXCLUDED.error, segments = EXCLUDED.segments, created_at = EXCLUDED.created_at`,
		plan.FilmID, plan.Status, plan.Error, segments, plan.CreatedAt)
	if err != nil {
		return fmt.Errorf("save plan: %w", err)
	}
	return nil
}

func (s *Store) ListPlannedFilmIDs(ctx context.Context, status string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT film_id FROM workout_plans WHERE $1 = '' OR status = $1`, status)
	if err != nil {
		return nil, fmt.Errorf("list planned films: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Store) ListStaleFailedPlanFilmIDs(ctx context.Context, before time.Time) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT film_id FROM workout_plans WHERE status = $1 AND created_at < $2`,
		workout.PlanStatusFailed, before)
	if err != nil {
		return nil, fmt.Errorf("list stale failed plans: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

const lessonColumns = `id, account_id, number, COALESCE(film_id::text, ''), start_ms, end_ms, status, steps, created_at, completed_at`

func (s *Store) CurrentLesson(ctx context.Context, accountID string) (workout.Lesson, error) {
	return scanLesson(s.pool.QueryRow(ctx, `SELECT `+lessonColumns+` FROM workout_lessons
		WHERE account_id = $1 AND status = $2 ORDER BY number DESC LIMIT 1`, accountID, workout.LessonStatusActive))
}

func (s *Store) PendingLesson(ctx context.Context, accountID string) (workout.Lesson, error) {
	return scanLesson(s.pool.QueryRow(ctx, `SELECT `+lessonColumns+` FROM workout_lessons
		WHERE account_id = $1 AND status IN ($2, $3) ORDER BY created_at DESC LIMIT 1`,
		accountID, workout.LessonStatusGenerating, workout.LessonStatusFailed))
}

func (s *Store) GetLesson(ctx context.Context, accountID, id string) (workout.Lesson, error) {
	return scanLesson(s.pool.QueryRow(ctx, `SELECT `+lessonColumns+` FROM workout_lessons
		WHERE account_id = $1 AND id = $2`, accountID, id))
}

func (s *Store) ListRecentLessons(ctx context.Context, accountID string, limit int) ([]workout.Lesson, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+lessonColumns+` FROM workout_lessons
		WHERE account_id = $1 AND status IN ('active','completed') ORDER BY number DESC LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	defer rows.Close()
	out := []workout.Lesson{}
	for rows.Next() {
		lesson, err := scanLesson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, lesson)
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanLesson(row scannable) (workout.Lesson, error) {
	var lesson workout.Lesson
	var steps []byte
	err := row.Scan(&lesson.ID, &lesson.AccountID, &lesson.Number, &lesson.FilmID, &lesson.StartMs, &lesson.EndMs,
		&lesson.Status, &steps, &lesson.CreatedAt, &lesson.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return workout.Lesson{}, shared.ErrNotFound
	}
	if err != nil {
		return workout.Lesson{}, fmt.Errorf("scan lesson: %w", err)
	}
	if err := json.Unmarshal(steps, &lesson.Steps); err != nil {
		return workout.Lesson{}, fmt.Errorf("unmarshal steps: %w", err)
	}
	return lesson, nil
}

func (s *Store) ClaimGeneration(ctx context.Context, lesson workout.Lesson, staleBefore time.Time) error {
	steps, err := json.Marshal(lesson.Steps)
	if err != nil {
		return fmt.Errorf("marshal steps: %w", err)
	}
	// Reclaim failed/abandoned rows first, then insert; both in one transaction
	// (a CTE would not work: INSERT does not see the DELETE within one statement).
	// A live claim still hits UNIQUE (account_id, number) and surfaces as a conflict.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("claim generation: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM workout_lessons
		WHERE account_id = $1 AND (status = 'failed' OR (status = 'generating' AND created_at < $2))`,
		lesson.AccountID, staleBefore); err != nil {
		return fmt.Errorf("claim generation: reclaim: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO workout_lessons (id, account_id, number, film_id, start_ms, end_ms, status, steps, created_at)
		VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,$8,$9)`,
		lesson.ID, lesson.AccountID, lesson.Number, lesson.FilmID, lesson.StartMs, lesson.EndMs, lesson.Status, steps, lesson.CreatedAt)
	if postgres.IsUniqueViolation(err) {
		return fmt.Errorf("lesson generation already in progress: %w", shared.ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("claim generation: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Store) UpdateLesson(ctx context.Context, lesson workout.Lesson) error {
	steps, err := json.Marshal(lesson.Steps)
	if err != nil {
		return fmt.Errorf("marshal steps: %w", err)
	}
	ct, err := s.pool.Exec(ctx, `UPDATE workout_lessons SET status = $3, steps = $4, completed_at = $5,
		film_id = NULLIF($6,'')::uuid, start_ms = $7, end_ms = $8
		WHERE account_id = $1 AND id = $2`,
		lesson.AccountID, lesson.ID, lesson.Status, steps, lesson.CompletedAt, lesson.FilmID, lesson.StartMs, lesson.EndMs)
	if err != nil {
		return fmt.Errorf("update lesson: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) ListCompletedDates(ctx context.Context, accountID string, since time.Time) ([]time.Time, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT date_trunc('day', completed_at) AS day FROM workout_lessons
		WHERE account_id = $1 AND status = $2 AND completed_at >= $3 ORDER BY day DESC`,
		accountID, workout.LessonStatusCompleted, since)
	if err != nil {
		return nil, fmt.Errorf("list completed dates: %w", err)
	}
	defer rows.Close()
	out := []time.Time{}
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) ListAccountsNeedingLesson(ctx context.Context, staleBefore time.Time) ([]string, error) {
	// Every row is either completed or a dead claim: accounts with an active
	// lesson or a live generation claim are excluded.
	rows, err := s.pool.Query(ctx, `SELECT account_id FROM workout_lessons
		GROUP BY account_id
		HAVING bool_and(status = 'completed' OR status = 'failed' OR (status = 'generating' AND created_at < $1))`, staleBefore)
	if err != nil {
		return nil, fmt.Errorf("list accounts needing lesson: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Store) ListItems(ctx context.Context, accountID string, sinceLesson int) ([]workout.Item, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, account_id, kind, text, film_id, start_ms, end_ms, lesson_number, last_score, times_reviewed, last_lesson, updated_at
		FROM workout_items WHERE account_id = $1 AND lesson_number >= $2`, accountID, sinceLesson)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()
	out := []workout.Item{}
	for rows.Next() {
		var it workout.Item
		if err := rows.Scan(&it.ID, &it.AccountID, &it.Kind, &it.Text, &it.FilmID, &it.StartMs, &it.EndMs,
			&it.LessonNumber, &it.LastScore, &it.TimesReviewed, &it.LastLesson, &it.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) UpsertItems(ctx context.Context, accountID string, lessonNumber int, results []workout.ItemResult, now time.Time) error {
	for _, r := range results {
		_, err := s.pool.Exec(ctx, `INSERT INTO workout_items (id, account_id, kind, text, film_id, start_ms, end_ms, lesson_number, last_score, times_reviewed, last_lesson, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, 0, $7, $9)
			ON CONFLICT (account_id, kind, text) DO UPDATE SET last_score = EXCLUDED.last_score, updated_at = EXCLUDED.updated_at`,
			accountID, r.Kind, r.Text, r.FilmID, r.StartMs, r.EndMs, lessonNumber, r.Score, now)
		if err != nil {
			return fmt.Errorf("upsert item: %w", err)
		}
	}
	return nil
}

func (s *Store) MarkReviewed(ctx context.Context, accountID string, texts []string, lessonNumber int, now time.Time) error {
	if len(texts) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, `UPDATE workout_items SET times_reviewed = times_reviewed + 1, last_lesson = $3, updated_at = $4
		WHERE account_id = $1 AND text = ANY($2)`, accountID, texts, lessonNumber, now)
	if err != nil {
		return fmt.Errorf("mark reviewed: %w", err)
	}
	return nil
}

func (s *Store) ListPositions(ctx context.Context, accountID string) ([]workout.Position, error) {
	rows, err := s.pool.Query(ctx, `SELECT account_id, title, film_id, next_segment, used_at FROM workout_positions WHERE account_id = $1`, accountID)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}
	defer rows.Close()
	out := []workout.Position{}
	for rows.Next() {
		var p workout.Position
		if err := rows.Scan(&p.AccountID, &p.Title, &p.FilmID, &p.NextSegment, &p.UsedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) SavePosition(ctx context.Context, pos workout.Position) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO workout_positions (account_id, title, film_id, next_segment, used_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (account_id, title) DO UPDATE SET film_id = EXCLUDED.film_id, next_segment = EXCLUDED.next_segment, used_at = EXCLUDED.used_at`,
		pos.AccountID, pos.Title, pos.FilmID, pos.NextSegment, pos.UsedAt)
	if err != nil {
		return fmt.Errorf("save position: %w", err)
	}
	return nil
}

// FailStaleGenerating marks lessons left mid-generation (e.g. after a restart)
// as failed, so the UI stops showing an eternal spinner and allows a retry.
func (s *Store) FailStaleGenerating(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE workout_lessons SET status = 'failed' WHERE status = 'generating'`)
	if err != nil {
		return fmt.Errorf("fail stale generating lessons: %w", err)
	}
	return nil
}

func (s *Store) DeleteAccountData(ctx context.Context, accountID string) error {
	for _, q := range []string{
		`DELETE FROM workout_lessons WHERE account_id = $1`,
		`DELETE FROM workout_items WHERE account_id = $1`,
		`DELETE FROM workout_positions WHERE account_id = $1`,
	} {
		if _, err := s.pool.Exec(ctx, q, accountID); err != nil {
			return fmt.Errorf("reset workout data: %w", err)
		}
	}
	return nil
}

// ListRecentErrors reads the learner's recent grammar mistakes from the core events pipeline.
func (s *Store) ListRecentErrors(ctx context.Context, accountID string, since time.Time, limit int) ([]workout.GrammarFocus, error) {
	rows, err := s.pool.Query(ctx, `SELECT grammar_key, error FROM events
		WHERE user_id = $1 AND outcome = 'fail' AND error IS NOT NULL AND created_at >= $2
		ORDER BY created_at DESC LIMIT $3`, accountID, since, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent errors: %w", err)
	}
	defer rows.Close()
	out := []workout.GrammarFocus{}
	for rows.Next() {
		var rule string
		var errJSON []byte
		if err := rows.Scan(&rule, &errJSON); err != nil {
			return nil, err
		}
		focus := workout.GrammarFocus{Rule: rule}
		var detail struct {
			Name        string `json:"name"`
			Sentence    string `json:"sentence"`
			Description string `json:"description"`
		}
		if json.Unmarshal(errJSON, &detail) == nil {
			if focus.Rule == "" {
				focus.Rule = detail.Name
			}
			focus.Sentence = detail.Sentence
			focus.Hint = detail.Description
		}
		if focus.Rule == "" && focus.Sentence == "" {
			continue
		}
		out = append(out, focus)
	}
	return out, rows.Err()
}
