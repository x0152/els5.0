package reader

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared"
)

const (
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusFailed     = "failed"
)

const (
	KindBook    = "book"
	KindArticle = "article"
)

type Book struct {
	ID          string
	OwnerID     string
	Title       string
	Author      string
	Description string
	CoverPath   string
	ContentPath string
	TextLength  int
	Position    int
	Status      string
	Error       string
	Kind        string
	GroupTitle  string
	CreatedAt   time.Time
}

type Collection struct {
	Title       string
	Description string
	CoverPath   string
}

func (b Book) Validate() error {
	var errs []error
	if strings.TrimSpace(b.Title) == "" {
		errs = append(errs, fmt.Errorf("book.title: must not be empty"))
	}
	if strings.TrimSpace(b.OwnerID) == "" {
		errs = append(errs, fmt.Errorf("book.owner_id: must not be empty"))
	}
	return shared.Validation(errs...)
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

func TextLength(htmlContent string) int {
	text := html.UnescapeString(tagRe.ReplaceAllString(htmlContent, " "))
	return len([]rune(strings.TrimSpace(text)))
}
