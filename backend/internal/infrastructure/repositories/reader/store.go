package reader

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const columns = `b.id, b.owner_id, b.title, b.author, b.description, b.cover_path, b.content_path, b.text_length, COALESCE(p.position, 0), b.status, b.error, b.created_at, b.kind, b.group_title`

func (s *Store) Create(ctx context.Context, b reader.Book) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO reader_books (id, owner_id, title, author, description, cover_path, content_path, text_length, status, error, created_at, kind, group_title)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		b.ID, b.OwnerID, b.Title, b.Author, b.Description, b.CoverPath, b.ContentPath, b.TextLength, b.Status, b.Error, b.CreatedAt, kindOrDefault(b.Kind), b.GroupTitle)
	if err != nil {
		return fmt.Errorf("create book: %w", err)
	}
	return nil
}

func (s *Store) List(ctx context.Context, viewerID string) ([]reader.Book, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+columns+` FROM reader_books b
		LEFT JOIN reader_progress p ON p.book_id = b.id AND p.owner_id = $1
		ORDER BY b.created_at DESC`, viewerID)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()
	out := []reader.Book{}
	for rows.Next() {
		b, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, viewerID, id string) (reader.Book, error) {
	b, err := scan(s.pool.QueryRow(ctx, `SELECT `+columns+` FROM reader_books b
		LEFT JOIN reader_progress p ON p.book_id = b.id AND p.owner_id = $1
		WHERE b.id = $2`, viewerID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return reader.Book{}, shared.ErrNotFound
	}
	if err != nil {
		return reader.Book{}, fmt.Errorf("get book: %w", err)
	}
	return b, nil
}

func (s *Store) Update(ctx context.Context, b reader.Book) error {
	ct, err := s.pool.Exec(ctx, `UPDATE reader_books
		SET title=$2, author=$3, description=$4, cover_path=$5, content_path=$6, text_length=$7, status=$8, error=$9, kind=$10, group_title=$11
		WHERE id=$1`,
		b.ID, b.Title, b.Author, b.Description, b.CoverPath, b.ContentPath, b.TextLength, b.Status, b.Error, kindOrDefault(b.Kind), b.GroupTitle)
	if err != nil {
		return fmt.Errorf("update book: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) SavePosition(ctx context.Context, viewerID, id string, position int) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO reader_progress (owner_id, book_id, position) VALUES ($1, $2, $3)
		ON CONFLICT (owner_id, book_id) DO UPDATE SET position = EXCLUDED.position`, viewerID, id, position)
	if err != nil {
		return fmt.Errorf("save position: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM reader_books WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) ListCollections(ctx context.Context) ([]reader.Collection, error) {
	rows, err := s.pool.Query(ctx, `SELECT title, description, cover_path FROM reader_collections`)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()
	out := []reader.Collection{}
	for rows.Next() {
		var c reader.Collection
		if err := rows.Scan(&c.Title, &c.Description, &c.CoverPath); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCollection(ctx context.Context, title string) (reader.Collection, error) {
	var c reader.Collection
	err := s.pool.QueryRow(ctx, `SELECT title, description, cover_path FROM reader_collections WHERE title = $1`, title).
		Scan(&c.Title, &c.Description, &c.CoverPath)
	if errors.Is(err, pgx.ErrNoRows) {
		return reader.Collection{}, shared.ErrNotFound
	}
	if err != nil {
		return reader.Collection{}, fmt.Errorf("get collection: %w", err)
	}
	return c, nil
}

func (s *Store) UpsertCollection(ctx context.Context, c reader.Collection) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO reader_collections (title, description, cover_path) VALUES ($1, $2, $3)
		ON CONFLICT (title) DO UPDATE SET description = EXCLUDED.description, cover_path = EXCLUDED.cover_path`,
		c.Title, c.Description, c.CoverPath)
	if err != nil {
		return fmt.Errorf("upsert collection: %w", err)
	}
	return nil
}

func (s *Store) RenameCollection(ctx context.Context, oldTitle, newTitle string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("rename collection: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE reader_books SET group_title = $2 WHERE group_title = $1`, oldTitle, newTitle); err != nil {
		return fmt.Errorf("rename collection articles: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE reader_collections SET title = $2 WHERE title = $1`, oldTitle, newTitle); err != nil {
		return fmt.Errorf("rename collection meta: %w", err)
	}
	return tx.Commit(ctx)
}

func kindOrDefault(kind string) string {
	if kind == reader.KindArticle {
		return reader.KindArticle
	}
	return reader.KindBook
}

type scannable interface {
	Scan(dest ...any) error
}

func scan(row scannable) (reader.Book, error) {
	var b reader.Book
	if err := row.Scan(&b.ID, &b.OwnerID, &b.Title, &b.Author, &b.Description, &b.CoverPath, &b.ContentPath, &b.TextLength, &b.Position, &b.Status, &b.Error, &b.CreatedAt, &b.Kind, &b.GroupTitle); err != nil {
		return reader.Book{}, err
	}
	return b, nil
}
