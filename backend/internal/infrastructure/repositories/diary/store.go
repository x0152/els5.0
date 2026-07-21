package diary

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/domain/shared"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const entryColumns = `id, account_id, entry_date, question, draft, text, reply, next_question, native_sample, corrections, created_at`

func (s *Store) Insert(ctx context.Context, e diary.Entry) error {
	corrections, err := json.Marshal(e.Corrections)
	if err != nil {
		return fmt.Errorf("marshal corrections: %w", err)
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO diary_entries (id, account_id, entry_date, question, draft, text, reply, next_question, native_sample, corrections, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.AccountID, e.Date, e.Question, e.Draft, e.Text, e.Reply, e.NextQuestion, e.NativeSample, corrections, e.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert diary entry: %w", err)
	}
	return nil
}

func (s *Store) GetByDate(ctx context.Context, accountID string, date time.Time) (diary.Entry, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT `+entryColumns+` FROM diary_entries WHERE account_id = $1 AND entry_date = $2`,
		accountID, date)
	return scanEntry(row)
}

func (s *Store) List(ctx context.Context, accountID string, limit, offset int32) ([]diary.Entry, int64, error) {
	var total int64
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM diary_entries WHERE account_id = $1`, accountID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count diary entries: %w", err)
	}
	rows, err := s.pool.Query(ctx,
		`SELECT `+entryColumns+` FROM diary_entries WHERE account_id = $1 ORDER BY entry_date DESC LIMIT $2 OFFSET $3`,
		accountID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list diary entries: %w", err)
	}
	defer rows.Close()
	out := make([]diary.Entry, 0, limit)
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (s *Store) Latest(ctx context.Context, accountID string, n int32) ([]diary.Entry, error) {
	entries, _, err := s.List(ctx, accountID, n, 0)
	return entries, err
}

func (s *Store) Dates(ctx context.Context, accountID string, limit int32) ([]time.Time, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT entry_date FROM diary_entries WHERE account_id = $1 ORDER BY entry_date DESC LIMIT $2`,
		accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("list diary dates: %w", err)
	}
	defer rows.Close()
	out := make([]time.Time, 0, limit)
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("scan diary date: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) DeleteAll(ctx context.Context, accountID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM diary_entries WHERE account_id = $1`, accountID)
	if err != nil {
		return fmt.Errorf("delete diary entries: %w", err)
	}
	return nil
}

func scanEntry(row pgx.Row) (diary.Entry, error) {
	var e diary.Entry
	var corrections []byte
	err := row.Scan(&e.ID, &e.AccountID, &e.Date, &e.Question, &e.Draft, &e.Text, &e.Reply, &e.NextQuestion, &e.NativeSample, &corrections, &e.CreatedAt)
	if err == pgx.ErrNoRows {
		return diary.Entry{}, shared.ErrNotFound
	}
	if err != nil {
		return diary.Entry{}, fmt.Errorf("scan diary entry: %w", err)
	}
	if err := json.Unmarshal(corrections, &e.Corrections); err != nil {
		return diary.Entry{}, fmt.Errorf("unmarshal corrections: %w", err)
	}
	return e, nil
}
