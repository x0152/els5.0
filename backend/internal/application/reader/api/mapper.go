package api

import (
	"time"

	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/reader"
)

func percent(book reader.Book) int {
	if book.TextLength <= 0 {
		return 0
	}
	p := book.Position * 100 / book.TextLength
	if p > 100 {
		return 100
	}
	return p
}

func kindOrDefault(kind string) string {
	if kind == reader.KindArticle {
		return reader.KindArticle
	}
	return reader.KindBook
}

func toBookSummary(book reader.Book, urls media.PublicURL) BookSummary {
	return BookSummary{
		ID:          book.ID,
		Title:       book.Title,
		Author:      book.Author,
		Description: book.Description,
		CoverURL:    buildURL(urls, book.CoverPath),
		Status:      book.Status,
		Kind:        kindOrDefault(book.Kind),
		GroupTitle:  book.GroupTitle,
		TextLength:  book.TextLength,
		Position:    book.Position,
		Percent:     percent(book),
		CreatedAt:   book.CreatedAt.Format(time.RFC3339),
	}
}

func toBooksOutput(list []reader.Book, urls media.PublicURL) BooksOutput {
	items := make([]BookSummary, 0, len(list))
	for _, b := range list {
		items = append(items, toBookSummary(b, urls))
	}
	return BooksOutput{Items: items}
}

func buildURL(urls media.PublicURL, raw string) string {
	if raw == "" {
		return ""
	}
	return urls.BuildFromRaw(raw)
}

func toBookOutput(book reader.Book, urls media.PublicURL) BookOutput {
	return BookOutput{
		ID:          book.ID,
		Title:       book.Title,
		Author:      book.Author,
		Description: book.Description,
		CoverURL:    buildURL(urls, book.CoverPath),
		Status:      book.Status,
		Error:       book.Error,
		Kind:        kindOrDefault(book.Kind),
		GroupTitle:  book.GroupTitle,
		TextLength:  book.TextLength,
		Position:    book.Position,
		Percent:     percent(book),
		ContentURL:  buildURL(urls, book.ContentPath),
		CreatedAt:   book.CreatedAt.Format(time.RFC3339),
	}
}

func toCollectionOutput(c reader.Collection, urls media.PublicURL) CollectionOutput {
	return CollectionOutput{Title: c.Title, Description: c.Description, CoverURL: buildURL(urls, c.CoverPath)}
}

func toCollectionsOutput(list []reader.Collection, urls media.PublicURL) CollectionsOutput {
	items := make([]CollectionOutput, 0, len(list))
	for _, c := range list {
		items = append(items, toCollectionOutput(c, urls))
	}
	return CollectionsOutput{Items: items}
}
