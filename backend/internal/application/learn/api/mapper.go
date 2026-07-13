package api

import (
	"strings"

	usecases "github.com/els/backend/internal/application/learn/use_cases"
	"github.com/els/backend/internal/domain/book"
	"github.com/els/backend/internal/domain/illustration"
	"github.com/els/backend/internal/domain/practice"
)

// --- books ---

func toBookListOutput(books []book.Book) BookListOutput {
	items := make([]BookSchema, 0, len(books))
	for _, b := range books {
		items = append(items, BookSchema{Slug: b.Slug, Series: b.Series, Level: b.Level, Title: b.Title, Description: b.Description})
	}
	return BookListOutput{Items: items}
}

// --- chapters ---

func toChapterOutput(c book.Chapter) ChapterOutput {
	status := c.Status
	if status == "" {
		status = book.StatusReady
	}
	return ChapterOutput{
		Number:    c.Number,
		Title:     c.Title,
		Page:      c.Page,
		Words:     c.Words,
		Footer:    c.Footer,
		Theory:    c.Theory,
		Exercises: c.Exercises,
		Status:    status,
		Error:     c.Error,
	}
}

func toChaptersOutput(chapters []book.Chapter) ChaptersOutput {
	items := make([]ChapterOutput, 0, len(chapters))
	for _, c := range chapters {
		items = append(items, toChapterOutput(c))
	}
	return ChaptersOutput{Items: items}
}

func toChapter(bk string, body ChapterBody) book.Chapter {
	words := body.Words
	if words == nil {
		words = []string{}
	}
	return book.Chapter{
		Book:      bk,
		Number:    body.Number,
		Title:     strings.TrimSpace(body.Title),
		Page:      body.Page,
		Words:     words,
		Footer:    strings.TrimSpace(body.Footer),
		Theory:    body.Theory,
		Exercises: body.Exercises,
	}
}

// --- illustrations ---

func toEnsureCommand(body EnsureBody) usecases.EnsureIllustrationCommand {
	return usecases.EnsureIllustrationCommand{
		Prompt:  strings.TrimSpace(body.Prompt),
		Aspect:  body.Aspect,
		Trigger: body.Trigger,
	}
}

func toIllustrationOutput(status illustration.Status) IllustrationOutput {
	return IllustrationOutput{
		ID:     status.ID,
		Status: status.Status,
		URL:    status.URL,
		Error:  status.Error,
	}
}

// --- practice ---

func toVariantSchema(v practice.Variant) VariantSchema {
	return VariantSchema{ID: v.ID, Title: v.Title, Exercises: v.Exercises, Status: v.Status, Error: v.Error}
}

func toVariantsOutput(variants []practice.Variant) VariantsOutput {
	items := make([]VariantSchema, 0, len(variants))
	for _, v := range variants {
		items = append(items, toVariantSchema(v))
	}
	return VariantsOutput{Items: items}
}

func toProgressOutput(p practice.Progress) ProgressOutput {
	answers := make(map[string]AnswerSchema, len(p.Answers))
	for k, a := range p.Answers {
		answers[k] = AnswerSchema(a)
	}
	return ProgressOutput{Answers: answers, Completed: p.Completed}
}

func toProgress(body SaveProgressBody) practice.Progress {
	answers := make(map[string]practice.AnswerState, len(body.Answers))
	for k, a := range body.Answers {
		answers[k] = practice.AnswerState(a)
	}
	return practice.Progress{Answers: answers, Completed: body.Completed}
}

func toCheckOutput(r practice.CheckResult) CheckOutput {
	return CheckOutput{Correct: r.Correct, Correction: r.Correction, Explanation: r.Explanation}
}

func variantKey(v string) string {
	if v == "" {
		return practice.MainVariant
	}
	return v
}
