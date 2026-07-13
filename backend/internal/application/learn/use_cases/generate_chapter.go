package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/book"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

// ChapterGenerator runs full-chapter generation (theory + exercises) in the background.
type ChapterGenerator interface {
	Enqueue(book string, number int, topic string)
}

type GenerateChapterUseCase struct {
	chapters  book.Repository
	generator ChapterGenerator
}

func NewGenerateChapterUseCase(chapters book.Repository, generator ChapterGenerator) *GenerateChapterUseCase {
	return &GenerateChapterUseCase{chapters: chapters, generator: generator}
}

func (uc *GenerateChapterUseCase) Execute(ctx context.Context, actor *iam.Actor, bk, topic string) (book.Chapter, error) {
	// 1. Only a global admin fills the book.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return book.Chapter{}, err
	}

	// 2. Topic is required.
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return book.Chapter{}, shared.Validation(fmt.Errorf("topic: must not be empty"))
	}

	// 3. Assign the next number in the book.
	existing, err := uc.chapters.List(ctx, bk)
	if err != nil {
		return book.Chapter{}, err
	}
	next := 1
	for _, c := range existing {
		if c.Number >= next {
			next = c.Number + 1
		}
	}

	// 4. Save a stub with status generating — the list shows it immediately; reload does not lose the job.
	chapter := book.Chapter{
		ID:     uuid.NewString(),
		Book:   bk,
		Number: next,
		Title:  topic,
		Status: book.StatusGenerating,
	}
	if err := chapter.Validate(); err != nil {
		return book.Chapter{}, err
	}
	if err := uc.chapters.Create(ctx, chapter); err != nil {
		return book.Chapter{}, err
	}

	// 5. Generation itself runs in the background; status is read via chapter-list polling.
	uc.generator.Enqueue(bk, next, topic)
	return chapter, nil
}
