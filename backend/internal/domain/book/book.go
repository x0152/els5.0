package book

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/els/backend/internal/domain/shared"
)

// Book is a collection of chapters (e.g. a grammar or vocabulary book).
// Books with the same Series are alternatives of one topic at different levels.
type Book struct {
	Slug        string
	Series      string
	Level       string
	Title       string
	Description string
}

const descriptionMarker = "---DESCRIPTION---"

// ParseBook parses a book.md seed file: "key: value" header lines,
// then an optional markdown description after ---DESCRIPTION---.
func ParseBook(slug, raw string) Book {
	b := Book{Slug: slug, Series: slug}
	header := raw
	if at := strings.Index(raw, descriptionMarker); at >= 0 {
		header = raw[:at]
		b.Description = strings.TrimSpace(raw[at+len(descriptionMarker):])
	}
	for _, line := range strings.Split(header, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "title":
			b.Title = value
		case "series":
			b.Series = value
		case "level":
			b.Level = value
		}
	}
	return b
}

const (
	StatusGenerating = "generating"
	StatusReady      = "ready"
	StatusError      = "error"
)

// Chapter is one page of a book: shared shape for every book kind.
type Chapter struct {
	ID        string
	Book      string
	Number    int
	Title     string
	Page      int
	Words     []string
	Footer    string
	Theory    string
	Exercises string
	Status    string
	Error     string
}

func (c Chapter) Validate() error {
	var errs []error
	if strings.TrimSpace(c.Book) == "" {
		errs = append(errs, fmt.Errorf("chapter.book: must not be empty"))
	}
	if c.Number <= 0 {
		errs = append(errs, fmt.Errorf("chapter.number: must be > 0"))
	}
	if strings.TrimSpace(c.Title) == "" {
		errs = append(errs, fmt.Errorf("chapter.title: must not be empty"))
	}
	return shared.Validation(errs...)
}

const (
	theoryMarker    = "---THEORY---"
	exercisesMarker = "---EXERCISES---"
)

func ParseChapter(book, raw string) (Chapter, error) {
	theoryAt := strings.Index(raw, theoryMarker)
	exercisesAt := strings.Index(raw, exercisesMarker)
	if theoryAt < 0 || exercisesAt < 0 || exercisesAt < theoryAt {
		return Chapter{}, shared.Validation(fmt.Errorf("chapter: missing %s/%s markers", theoryMarker, exercisesMarker))
	}

	c := Chapter{
		Book:      book,
		Words:     []string{},
		Theory:    strings.TrimSpace(raw[theoryAt+len(theoryMarker) : exercisesAt]),
		Exercises: strings.TrimSpace(raw[exercisesAt+len(exercisesMarker):]),
	}
	for _, line := range strings.Split(raw[:theoryAt], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "number":
			c.Number, _ = strconv.Atoi(value)
		case "title":
			c.Title = value
		case "page":
			c.Page, _ = strconv.Atoi(value)
		case "footer":
			c.Footer = value
		case "words":
			c.Words = splitList(value)
		}
	}
	if err := c.Validate(); err != nil {
		return Chapter{}, err
	}
	return c, nil
}

func splitList(value string) []string {
	out := []string{}
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
