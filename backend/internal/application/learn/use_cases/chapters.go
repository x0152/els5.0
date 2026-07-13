package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/book"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type ListBooksUseCase struct {
	chapters book.Repository
}

func NewListBooksUseCase(chapters book.Repository) *ListBooksUseCase {
	return &ListBooksUseCase{chapters: chapters}
}

func (uc *ListBooksUseCase) Execute(ctx context.Context, actor *iam.Actor) ([]book.Book, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	return uc.chapters.ListBooks(ctx)
}

type ListChaptersUseCase struct {
	chapters book.Repository
}

func NewListChaptersUseCase(chapters book.Repository) *ListChaptersUseCase {
	return &ListChaptersUseCase{chapters: chapters}
}

func (uc *ListChaptersUseCase) Execute(ctx context.Context, actor *iam.Actor, bk string) ([]book.Chapter, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	return uc.chapters.List(ctx, bk)
}

type GetChapterUseCase struct {
	chapters book.Repository
}

func NewGetChapterUseCase(chapters book.Repository) *GetChapterUseCase {
	return &GetChapterUseCase{chapters: chapters}
}

func (uc *GetChapterUseCase) Execute(ctx context.Context, actor *iam.Actor, bk string, number int) (book.Chapter, error) {
	if actor == nil {
		return book.Chapter{}, shared.ErrUnauthorized
	}
	return uc.chapters.GetByNumber(ctx, bk, number)
}

type CreateChapterUseCase struct {
	chapters book.Repository
}

func NewCreateChapterUseCase(chapters book.Repository) *CreateChapterUseCase {
	return &CreateChapterUseCase{chapters: chapters}
}

func (uc *CreateChapterUseCase) Execute(ctx context.Context, actor *iam.Actor, chapter book.Chapter) (book.Chapter, error) {
	// 1. Only a global admin edits the book.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return book.Chapter{}, err
	}

	// 2. The chapter validates its own invariants.
	if err := chapter.Validate(); err != nil {
		return book.Chapter{}, err
	}

	// 3. Persist the chapter with a new id.
	chapter.ID = uuid.NewString()
	if err := uc.chapters.Create(ctx, chapter); err != nil {
		return book.Chapter{}, err
	}
	return chapter, nil
}

type UpdateChapterUseCase struct {
	chapters book.Repository
}

func NewUpdateChapterUseCase(chapters book.Repository) *UpdateChapterUseCase {
	return &UpdateChapterUseCase{chapters: chapters}
}

func (uc *UpdateChapterUseCase) Execute(ctx context.Context, actor *iam.Actor, chapter book.Chapter) (book.Chapter, error) {
	// 1. Only a global admin edits the book.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return book.Chapter{}, err
	}

	// 2. The chapter validates its own invariants.
	if err := chapter.Validate(); err != nil {
		return book.Chapter{}, err
	}

	// 3. Update the chapter by book and number.
	if err := uc.chapters.Update(ctx, chapter); err != nil {
		return book.Chapter{}, err
	}
	return chapter, nil
}

type DeleteChapterUseCase struct {
	chapters book.Repository
}

func NewDeleteChapterUseCase(chapters book.Repository) *DeleteChapterUseCase {
	return &DeleteChapterUseCase{chapters: chapters}
}

func (uc *DeleteChapterUseCase) Execute(ctx context.Context, actor *iam.Actor, bk string, number int) error {
	// 1. Only a global admin edits the book.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return err
	}

	// 2. Delete the chapter by book and number.
	return uc.chapters.DeleteByNumber(ctx, bk, number)
}
