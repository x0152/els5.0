package practice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/practice"
	"github.com/els/backend/internal/domain/shared"
)

type VariantStore struct {
	pool *pgxpool.Pool
}

func NewVariantStore(pool *pgxpool.Pool) *VariantStore { return &VariantStore{pool: pool} }

func (s *VariantStore) List(ctx context.Context, accountID string, kind practice.Kind, number int) ([]practice.Variant, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, title, exercises, status, error FROM practice_variants
		WHERE account_id=$1 AND kind=$2 AND number=$3 ORDER BY created_at`, accountID, string(kind), number)
	if err != nil {
		return nil, fmt.Errorf("list variants: %w", err)
	}
	defer rows.Close()
	out := []practice.Variant{}
	for rows.Next() {
		v := practice.Variant{Kind: kind, Number: number}
		if err := rows.Scan(&v.ID, &v.Title, &v.Exercises, &v.Status, &v.Error); err != nil {
			return nil, fmt.Errorf("scan variant: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *VariantStore) Get(ctx context.Context, accountID, id string) (practice.Variant, error) {
	var v practice.Variant
	var kind string
	err := s.pool.QueryRow(ctx, `SELECT id, kind, number, title, exercises, status, error FROM practice_variants
		WHERE account_id=$1 AND id=$2`, accountID, id).Scan(&v.ID, &kind, &v.Number, &v.Title, &v.Exercises, &v.Status, &v.Error)
	if errors.Is(err, pgx.ErrNoRows) {
		return practice.Variant{}, shared.ErrNotFound
	}
	if err != nil {
		return practice.Variant{}, fmt.Errorf("get variant: %w", err)
	}
	v.Kind = practice.Kind(kind)
	return v, nil
}

func (s *VariantStore) Create(ctx context.Context, accountID string, v practice.Variant) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO practice_variants (id, account_id, kind, number, title, exercises, status, error, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now())`, v.ID, accountID, string(v.Kind), v.Number, v.Title, v.Exercises, v.Status, v.Error)
	if err != nil {
		return fmt.Errorf("create variant: %w", err)
	}
	return nil
}

func (s *VariantStore) Update(ctx context.Context, accountID string, v practice.Variant) error {
	ct, err := s.pool.Exec(ctx, `UPDATE practice_variants SET title=$1, exercises=$2, status=$3, error=$4
		WHERE account_id=$5 AND id=$6`, v.Title, v.Exercises, v.Status, v.Error, accountID, v.ID)
	if err != nil {
		return fmt.Errorf("update variant: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// FailStaleGenerating marks variants left mid-generation (e.g. after a restart) as failed.
func (s *VariantStore) FailStaleGenerating(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE practice_variants SET status=$1, error='generation interrupted' WHERE status=$2`,
		practice.StatusError, practice.StatusGenerating)
	if err != nil {
		return fmt.Errorf("fail stale variants: %w", err)
	}
	return nil
}

func (s *VariantStore) Delete(ctx context.Context, accountID, id string) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM practice_variants WHERE account_id=$1 AND id=$2`, accountID, id)
	if err != nil {
		return fmt.Errorf("delete variant: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

type ProgressStore struct {
	pool *pgxpool.Pool
}

func NewProgressStore(pool *pgxpool.Pool) *ProgressStore { return &ProgressStore{pool: pool} }

func (s *ProgressStore) Get(ctx context.Context, accountID string, kind practice.Kind, number int, variantKey string) (practice.Progress, error) {
	var answers []byte
	var completed bool
	err := s.pool.QueryRow(ctx, `SELECT answers, completed FROM practice_progress
		WHERE account_id=$1 AND kind=$2 AND number=$3 AND variant_key=$4`,
		accountID, string(kind), number, variantKey).Scan(&answers, &completed)
	if errors.Is(err, pgx.ErrNoRows) {
		return practice.Progress{Answers: map[string]practice.AnswerState{}}, nil
	}
	if err != nil {
		return practice.Progress{}, fmt.Errorf("get progress: %w", err)
	}
	m := map[string]practice.AnswerState{}
	if len(answers) > 0 {
		_ = json.Unmarshal(answers, &m)
	}
	return practice.Progress{Answers: m, Completed: completed}, nil
}

func (s *ProgressStore) Save(ctx context.Context, accountID string, kind practice.Kind, number int, variantKey string, p practice.Progress) error {
	payload, err := json.Marshal(p.Answers)
	if err != nil {
		return fmt.Errorf("marshal answers: %w", err)
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO practice_progress (account_id, kind, number, variant_key, answers, completed, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,now())
		ON CONFLICT (account_id, kind, number, variant_key)
		DO UPDATE SET answers=EXCLUDED.answers, completed=EXCLUDED.completed, updated_at=now()`,
		accountID, string(kind), number, variantKey, payload, p.Completed)
	if err != nil {
		return fmt.Errorf("save progress: %w", err)
	}
	return nil
}

func (s *ProgressStore) Delete(ctx context.Context, accountID string, kind practice.Kind, number int, variantKey string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM practice_progress
		WHERE account_id=$1 AND kind=$2 AND number=$3 AND variant_key=$4`,
		accountID, string(kind), number, variantKey)
	if err != nil {
		return fmt.Errorf("delete progress: %w", err)
	}
	return nil
}
