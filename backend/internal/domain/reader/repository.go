package reader

import "context"

type Repository interface {
	Create(ctx context.Context, book Book) error
	List(ctx context.Context, viewerID string) ([]Book, error)
	Get(ctx context.Context, viewerID, id string) (Book, error)
	Update(ctx context.Context, book Book) error
	SavePosition(ctx context.Context, viewerID, id string, position int) error
	Delete(ctx context.Context, id string) error

	ListCollections(ctx context.Context) ([]Collection, error)
	GetCollection(ctx context.Context, title string) (Collection, error)
	UpsertCollection(ctx context.Context, c Collection) error
	RenameCollection(ctx context.Context, oldTitle, newTitle string) error
}
