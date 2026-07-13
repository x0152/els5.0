package vocab

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
	"github.com/els/backend/internal/infrastructure/postgres"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const columns = `id, account_id, text, kind, transcription, translation, definition, example, frequency, cefr, status, created_at`

func (s *Store) Create(ctx context.Context, u vocab.Unit) (vocab.Unit, error) {
	err := s.pool.QueryRow(ctx, `INSERT INTO vocab_units (id, account_id, text, kind, transcription, translation, definition, example, frequency, cefr, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now())
		RETURNING created_at`,
		u.ID, u.AccountID, u.Text, string(u.Kind), u.Transcription, u.Translation, u.Definition, u.Example, u.Frequency, u.CEFR, string(u.Status)).Scan(&u.CreatedAt)
	if postgres.IsUniqueViolation(err) {
		return vocab.Unit{}, shared.ErrConflict
	}
	if err != nil {
		return vocab.Unit{}, fmt.Errorf("create unit: %w", err)
	}
	return u, nil
}

func (s *Store) List(ctx context.Context, accountID string, f vocab.ListFilter) ([]vocab.Unit, int, error) {
	where := []string{"account_id = $1"}
	args := []any{accountID}
	if f.Status != "" {
		args = append(args, string(f.Status))
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		where = append(where, fmt.Sprintf("(lower(text) LIKE $%d OR lower(translation) LIKE $%d)", len(args), len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM vocab_units WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count units: %w", err)
	}

	args = append(args, f.Limit, f.Offset)
	query := fmt.Sprintf(`SELECT %s FROM vocab_units WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		columns, clause, len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list units: %w", err)
	}
	defer rows.Close()
	out := []vocab.Unit{}
	for rows.Next() {
		unit, err := scanUnit(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, unit)
	}
	return out, total, rows.Err()
}

func (s *Store) ExistsText(ctx context.Context, accountID, text string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vocab_units WHERE account_id = $1 AND lower(text) = lower($2))`,
		accountID, text).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists unit: %w", err)
	}
	return exists, nil
}

func (s *Store) UpdateStatus(ctx context.Context, accountID, id string, status vocab.Status) (vocab.Unit, error) {
	unit, err := scanUnit(s.pool.QueryRow(ctx, `UPDATE vocab_units SET status = $1 WHERE id = $2 AND account_id = $3
		RETURNING `+columns, string(status), id, accountID))
	if errors.Is(err, pgx.ErrNoRows) {
		return vocab.Unit{}, shared.ErrNotFound
	}
	if err != nil {
		return vocab.Unit{}, fmt.Errorf("update unit status: %w", err)
	}
	return unit, nil
}

func (s *Store) Delete(ctx context.Context, accountID, id string) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM vocab_units WHERE id = $1 AND account_id = $2`, id, accountID)
	if err != nil {
		return fmt.Errorf("delete unit: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUnit(row scannable) (vocab.Unit, error) {
	var u vocab.Unit
	var kind, status string
	if err := row.Scan(&u.ID, &u.AccountID, &u.Text, &kind, &u.Transcription, &u.Translation, &u.Definition, &u.Example, &u.Frequency, &u.CEFR, &status, &u.CreatedAt); err != nil {
		return vocab.Unit{}, err
	}
	u.Kind = vocab.Kind(kind)
	u.Status = vocab.Status(status)
	return u, nil
}
