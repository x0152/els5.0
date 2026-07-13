package book

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/book"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/infrastructure/postgres"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const columns = `id, book, number, title, page, words, footer, theory, exercises, status, error`

func (s *Store) EnsureBook(ctx context.Context, b book.Book) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO books (slug, series, level, title, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,now(),now())
		ON CONFLICT (slug) DO UPDATE SET series=$2, level=$3, title=$4, description=$5, updated_at=now()`,
		b.Slug, b.Series, b.Level, b.Title, b.Description)
	if err != nil {
		return fmt.Errorf("ensure book: %w", err)
	}
	return nil
}

func (s *Store) ListBooks(ctx context.Context) ([]book.Book, error) {
	rows, err := s.pool.Query(ctx, `SELECT slug, series, level, title, description FROM books ORDER BY series, level, slug`)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()
	out := []book.Book{}
	for rows.Next() {
		var b book.Book
		if err := rows.Scan(&b.Slug, &b.Series, &b.Level, &b.Title, &b.Description); err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) Count(ctx context.Context, bk string) (int, error) {
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM book_chapters WHERE book = $1`, bk).Scan(&n); err != nil {
		return 0, fmt.Errorf("count chapters: %w", err)
	}
	return n, nil
}

func (s *Store) List(ctx context.Context, bk string) ([]book.Chapter, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+columns+` FROM book_chapters WHERE book = $1 ORDER BY number`, bk)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()
	out := []book.Chapter{}
	for rows.Next() {
		chapter, err := scanChapter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, chapter)
	}
	return out, rows.Err()
}

func (s *Store) GetByNumber(ctx context.Context, bk string, number int) (book.Chapter, error) {
	chapter, err := scanChapter(s.pool.QueryRow(ctx, `SELECT `+columns+` FROM book_chapters WHERE book = $1 AND number = $2`, bk, number))
	if errors.Is(err, pgx.ErrNoRows) {
		return book.Chapter{}, shared.ErrNotFound
	}
	if err != nil {
		return book.Chapter{}, fmt.Errorf("get chapter: %w", err)
	}
	return chapter, nil
}

func (s *Store) Create(ctx context.Context, c book.Chapter) error {
	list, err := json.Marshal(c.Words)
	if err != nil {
		return fmt.Errorf("marshal words: %w", err)
	}
	status := c.Status
	if status == "" {
		status = book.StatusReady
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO book_chapters (id, book, number, title, page, words, footer, theory, exercises, status, error, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now(),now())`,
		c.ID, c.Book, c.Number, c.Title, c.Page, list, c.Footer, c.Theory, c.Exercises, status, c.Error)
	if postgres.IsUniqueViolation(err) {
		return shared.ErrConflict
	}
	if err != nil {
		return fmt.Errorf("create chapter: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, c book.Chapter) error {
	list, err := json.Marshal(c.Words)
	if err != nil {
		return fmt.Errorf("marshal words: %w", err)
	}
	ct, err := s.pool.Exec(ctx, `UPDATE book_chapters SET title=$1, page=$2, words=$3, footer=$4, theory=$5, exercises=$6, status='ready', error='', updated_at=now()
		WHERE book=$7 AND number=$8`,
		c.Title, c.Page, list, c.Footer, c.Theory, c.Exercises, c.Book, c.Number)
	if err != nil {
		return fmt.Errorf("update chapter: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// FailStaleGenerating marks chapters left mid-generation (e.g. after a restart) as failed.
func (s *Store) FailStaleGenerating(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE book_chapters SET status='error', error='generation interrupted', updated_at=now() WHERE status='generating'`)
	if err != nil {
		return fmt.Errorf("fail stale chapters: %w", err)
	}
	return nil
}

func (s *Store) SetStatus(ctx context.Context, bk string, number int, status, errMsg string) error {
	_, err := s.pool.Exec(ctx, `UPDATE book_chapters SET status=$1, error=$2, updated_at=now() WHERE book=$3 AND number=$4`,
		status, errMsg, bk, number)
	if err != nil {
		return fmt.Errorf("set chapter status: %w", err)
	}
	return nil
}

func (s *Store) DeleteByNumber(ctx context.Context, bk string, number int) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM book_chapters WHERE book=$1 AND number=$2`, bk, number)
	if err != nil {
		return fmt.Errorf("delete chapter: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanChapter(row scannable) (book.Chapter, error) {
	var c book.Chapter
	var list []byte
	if err := row.Scan(&c.ID, &c.Book, &c.Number, &c.Title, &c.Page, &list, &c.Footer, &c.Theory, &c.Exercises, &c.Status, &c.Error); err != nil {
		return book.Chapter{}, err
	}
	c.Words = []string{}
	if len(list) > 0 {
		if err := json.Unmarshal(list, &c.Words); err != nil {
			return book.Chapter{}, fmt.Errorf("unmarshal words: %w", err)
		}
	}
	return c, nil
}
