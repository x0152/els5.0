package usecases

import (
	"bytes"
	"context"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type UpdateBookUseCase struct {
	books   reader.Repository
	storage media.Storage
	bucket  string
}

func NewUpdateBookUseCase(repo reader.Repository, storage media.Storage, bucket string) *UpdateBookUseCase {
	return &UpdateBookUseCase{books: repo, storage: storage, bucket: bucket}
}

type UpdateBookCommand struct {
	Title       string
	Author      string
	Description string
	Cover       *UploadAsset
}

func (uc *UpdateBookUseCase) Execute(ctx context.Context, actor *iam.Actor, id string, cmd UpdateBookCommand) (reader.Book, error) {
	if actor == nil {
		return reader.Book{}, shared.ErrUnauthorized
	}

	book, err := uc.books.Get(ctx, actor.AccountID().String(), id)
	if err != nil {
		return reader.Book{}, err
	}

	book.Title = strings.TrimSpace(cmd.Title)
	book.Author = strings.TrimSpace(cmd.Author)
	book.Description = strings.TrimSpace(cmd.Description)
	if err := book.Validate(); err != nil {
		return reader.Book{}, err
	}

	if cmd.Cover != nil && len(cmd.Cover.Data) > 0 {
		path, err := media.NewPath(uc.bucket + "/" + book.ID + "/cover" + ext(cmd.Cover.Filename, ".jpg"))
		if err != nil {
			return reader.Book{}, err
		}
		if err := uc.storage.Put(ctx, path, bytes.NewReader(cmd.Cover.Data), media.PutOptions{ContentType: cmd.Cover.ContentType, Size: int64(len(cmd.Cover.Data))}); err != nil {
			return reader.Book{}, err
		}
		if book.CoverPath != "" && book.CoverPath != path.String() {
			if old, err := media.NewPath(book.CoverPath); err == nil {
				_ = uc.storage.Delete(ctx, old)
			}
		}
		book.CoverPath = path.String()
	}

	if err := uc.books.Update(ctx, book); err != nil {
		return reader.Book{}, err
	}
	return book, nil
}
