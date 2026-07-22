package studio

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/studio"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) ListAreas(ctx context.Context, accountID string) ([]studio.AreaStats, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT a.id, a.account_id, a.title, a.icon, a.created_at,
		        count(i.id),
		        count(i.id) FILTER (WHERE i.listened AND i.spoken AND i.written AND i.recalled),
		        count(i.id) FILTER (WHERE i.next_review_at <= now())
		 FROM studio_areas a
		 LEFT JOIN studio_items i ON i.area_id = a.id
		 WHERE a.account_id = $1
		 GROUP BY a.id
		 ORDER BY a.created_at`,
		accountID)
	if err != nil {
		return nil, fmt.Errorf("list studio areas: %w", err)
	}
	defer rows.Close()
	out := make([]studio.AreaStats, 0)
	for rows.Next() {
		var a studio.AreaStats
		if err := rows.Scan(&a.ID, &a.AccountID, &a.Title, &a.Icon, &a.CreatedAt, &a.Total, &a.Done, &a.Due); err != nil {
			return nil, fmt.Errorf("scan studio area: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) InsertArea(ctx context.Context, a studio.Area) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO studio_areas (id, account_id, title, icon, created_at) VALUES ($1,$2,$3,$4,$5)`,
		a.ID, a.AccountID, a.Title, a.Icon, a.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert studio area: %w", err)
	}
	return nil
}

func (s *Store) DeleteArea(ctx context.Context, accountID, id string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM studio_areas WHERE account_id = $1 AND id = $2`, accountID, id)
	if err != nil {
		return fmt.Errorf("delete studio area: %w", err)
	}
	return nil
}

const itemColumns = `id, area_id, account_id, text, transcription, translation, explanation, explanation_native, example, task, listened, spoken, written, recalled, review_stage, next_review_at, created_at`

func (s *Store) ListItems(ctx context.Context, accountID, areaID string) ([]studio.Item, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+itemColumns+` FROM studio_items WHERE account_id = $1 AND area_id = $2 ORDER BY created_at`,
		accountID, areaID)
	if err != nil {
		return nil, fmt.Errorf("list studio items: %w", err)
	}
	defer rows.Close()
	out := make([]studio.Item, 0)
	for rows.Next() {
		i, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, accountID, id string) (studio.Item, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT `+itemColumns+` FROM studio_items WHERE account_id = $1 AND id = $2`, accountID, id)
	return scanItem(row)
}

func (s *Store) Insert(ctx context.Context, i studio.Item) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO studio_items (`+itemColumns+`)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		i.ID, i.AreaID, i.AccountID, i.Text, i.Transcription, i.Translation, i.Explanation, i.ExplanationNative, i.Example, i.Task, i.Listened, i.Spoken, i.Written, i.Recalled, i.ReviewStage, i.NextReviewAt, i.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert studio item: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, i studio.Item) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE studio_items SET transcription=$3, translation=$4, explanation=$5, explanation_native=$6, example=$7, task=$8, listened=$9, spoken=$10, written=$11, recalled=$12, review_stage=$13, next_review_at=$14
		 WHERE account_id = $1 AND id = $2`,
		i.AccountID, i.ID, i.Transcription, i.Translation, i.Explanation, i.ExplanationNative, i.Example, i.Task, i.Listened, i.Spoken, i.Written, i.Recalled, i.ReviewStage, i.NextReviewAt)
	if err != nil {
		return fmt.Errorf("update studio item: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, accountID, id string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM studio_items WHERE account_id = $1 AND id = $2`, accountID, id)
	if err != nil {
		return fmt.Errorf("delete studio item: %w", err)
	}
	return nil
}

func scanItem(row pgx.Row) (studio.Item, error) {
	var i studio.Item
	err := row.Scan(&i.ID, &i.AreaID, &i.AccountID, &i.Text, &i.Transcription, &i.Translation, &i.Explanation, &i.ExplanationNative, &i.Example, &i.Task, &i.Listened, &i.Spoken, &i.Written, &i.Recalled, &i.ReviewStage, &i.NextReviewAt, &i.CreatedAt)
	if err == pgx.ErrNoRows {
		return studio.Item{}, shared.ErrNotFound
	}
	if err != nil {
		return studio.Item{}, fmt.Errorf("scan studio item: %w", err)
	}
	return i, nil
}
