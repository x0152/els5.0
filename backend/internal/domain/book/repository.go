package book

import "context"

type Repository interface {
	EnsureBook(ctx context.Context, b Book) error
	ListBooks(ctx context.Context) ([]Book, error)
	Count(ctx context.Context, book string) (int, error)
	List(ctx context.Context, book string) ([]Chapter, error)
	GetByNumber(ctx context.Context, book string, number int) (Chapter, error)
	Create(ctx context.Context, chapter Chapter) error
	Update(ctx context.Context, chapter Chapter) error
	SetStatus(ctx context.Context, book string, number int, status, errMsg string) error
	DeleteByNumber(ctx context.Context, book string, number int) error
}
