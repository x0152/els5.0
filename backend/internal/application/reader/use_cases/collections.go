package usecases

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type ListCollectionsUseCase struct {
	books reader.Repository
}

func NewListCollectionsUseCase(repo reader.Repository) *ListCollectionsUseCase {
	return &ListCollectionsUseCase{books: repo}
}

func (uc *ListCollectionsUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]reader.Collection, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	return uc.books.ListCollections(ctx)
}

type UpdateCollectionUseCase struct {
	books   reader.Repository
	storage media.Storage
	bucket  string
}

func NewUpdateCollectionUseCase(repo reader.Repository, storage media.Storage, bucket string) *UpdateCollectionUseCase {
	return &UpdateCollectionUseCase{books: repo, storage: storage, bucket: bucket}
}

type UpdateCollectionCommand struct {
	Title       string
	NewTitle    string
	Description string
	Cover       *UploadAsset
}

func (uc *UpdateCollectionUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd UpdateCollectionCommand) (reader.Collection, error) {
	if actor == nil {
		return reader.Collection{}, shared.ErrUnauthorized
	}

	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return reader.Collection{}, fmt.Errorf("%w: collection.title: must not be empty", shared.ErrValidation)
	}
	newTitle := strings.TrimSpace(cmd.NewTitle)
	if newTitle == "" {
		newTitle = title
	}

	existing, err := uc.books.GetCollection(ctx, title)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return reader.Collection{}, err
	}

	if newTitle != title {
		if err := uc.books.RenameCollection(ctx, title, newTitle); err != nil {
			return reader.Collection{}, err
		}
	}

	coverPath := existing.CoverPath
	if cmd.Cover != nil && len(cmd.Cover.Data) > 0 {
		path, err := media.NewPath(uc.bucket + "/collections/" + uuid.NewString() + ext(cmd.Cover.Filename, ".jpg"))
		if err != nil {
			return reader.Collection{}, err
		}
		if err := uc.storage.Put(ctx, path, bytes.NewReader(cmd.Cover.Data), media.PutOptions{ContentType: cmd.Cover.ContentType, Size: int64(len(cmd.Cover.Data))}); err != nil {
			return reader.Collection{}, err
		}
		if coverPath != "" {
			if old, err := media.NewPath(coverPath); err == nil {
				_ = uc.storage.Delete(ctx, old)
			}
		}
		coverPath = path.String()
	}

	collection := reader.Collection{Title: newTitle, Description: strings.TrimSpace(cmd.Description), CoverPath: coverPath}
	if err := uc.books.UpsertCollection(ctx, collection); err != nil {
		return reader.Collection{}, err
	}
	return collection, nil
}
