package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/els/backend/internal/domain/book"
)

type Chapters struct {
	chapters book.Repository
	service  *Service
	logger   *slog.Logger
}

func NewChapters(chapters book.Repository, service *Service, logger *slog.Logger) *Chapters {
	if logger == nil {
		logger = slog.Default()
	}
	return &Chapters{chapters: chapters, service: service, logger: logger}
}

func (c *Chapters) Enqueue(bk string, number int, topic string) {
	go c.run(bk, number, topic)
}

func (c *Chapters) run(bk string, number int, topic string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("learn: chapter generation panic", slog.String("book", bk), slog.Int("number", number), slog.Any("panic", r))
			c.fail(bk, number, fmt.Errorf("internal error during generation"))
		}
	}()

	chapter, err := c.service.GenerateChapter(ctx, bk, topic)
	if err != nil {
		c.fail(bk, number, err)
		return
	}
	chapter.Book = bk
	chapter.Number = number
	chapter.Status = book.StatusReady

	saveCtx, saveCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer saveCancel()
	if err := c.chapters.Update(saveCtx, chapter); err != nil {
		c.logger.Error("learn: save chapter failed", slog.String("book", bk), slog.Int("number", number), slog.String("err", err.Error()))
	}
}

func (c *Chapters) fail(bk string, number int, cause error) {
	c.logger.Warn("learn: chapter generation failed", slog.String("book", bk), slog.Int("number", number), slog.String("err", cause.Error()))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = c.chapters.SetStatus(ctx, bk, number, book.StatusError, cause.Error())
}
