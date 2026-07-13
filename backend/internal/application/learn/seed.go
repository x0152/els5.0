package learn

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/els/backend/internal/domain/book"
)

//go:embed seed
var seedFS embed.FS

func seedBooks() ([]book.Book, error) {
	dirs, err := fs.ReadDir(seedFS, "seed")
	if err != nil {
		return nil, err
	}
	books := make([]book.Book, 0, len(dirs))
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		raw, err := seedFS.ReadFile("seed/" + d.Name() + "/book.md")
		if err != nil {
			return nil, fmt.Errorf("seed/%s: missing book.md: %w", d.Name(), err)
		}
		books = append(books, book.ParseBook(d.Name(), string(raw)))
	}
	return books, nil
}

func seedChapters() ([]book.Chapter, error) {
	var paths []string
	err := fs.WalkDir(seedFS, "seed", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, "/book.md") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	chapters := make([]book.Chapter, 0, len(paths))
	for _, path := range paths {
		raw, err := seedFS.ReadFile(path)
		if err != nil {
			return nil, err
		}
		bk := strings.Split(strings.TrimPrefix(path, "seed/"), "/")[0]
		chapter, err := book.ParseChapter(bk, string(raw))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		chapters = append(chapters, chapter)
	}
	return chapters, nil
}
