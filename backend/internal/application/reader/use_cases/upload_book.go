package usecases

import (
	"bytes"
	"context"
	"fmt"
	gohtml "html"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/lexicon"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
)

type UploadBookUseCase struct {
	books     reader.Repository
	storage   media.Storage
	converter ports.BookConverter
	urls      media.PublicURL
	analyzer  lexicon.Analyzer
	lex       lexicon.Repository
	bucket    string
	tempDir   string
	logger    *slog.Logger
}

func NewUploadBookUseCase(repo reader.Repository, storage media.Storage, converter ports.BookConverter, urls media.PublicURL, analyzer lexicon.Analyzer, lex lexicon.Repository, bucket, tempDir string, logger *slog.Logger) *UploadBookUseCase {
	return &UploadBookUseCase{books: repo, storage: storage, converter: converter, urls: urls, analyzer: analyzer, lex: lex, bucket: bucket, tempDir: tempDir, logger: logger}
}

type UploadAsset struct {
	Data        []byte
	ContentType string
	Filename    string
}

type UploadBookCommand struct {
	Title       string
	Author      string
	Description string
	Filename    string
	TempPath    string
	Kind        string
	GroupTitle  string
	Cover       *UploadAsset
}

func (uc *UploadBookUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd UploadBookCommand) (reader.Book, error) {
	if actor == nil {
		_ = os.Remove(cmd.TempPath)
		return reader.Book{}, shared.ErrUnauthorized
	}

	kind := reader.KindBook
	if cmd.Kind == reader.KindArticle {
		kind = reader.KindArticle
	}
	book := reader.Book{
		ID:          uuid.NewString(),
		OwnerID:     actor.AccountID().String(),
		Title:       bookTitle(cmd),
		Author:      strings.TrimSpace(cmd.Author),
		Description: strings.TrimSpace(cmd.Description),
		Status:      reader.StatusProcessing,
		Kind:        kind,
		GroupTitle:  strings.TrimSpace(cmd.GroupTitle),
		CreatedAt:   time.Now().UTC(),
	}
	if err := book.Validate(); err != nil {
		_ = os.Remove(cmd.TempPath)
		return reader.Book{}, err
	}
	if err := uc.books.Create(ctx, book); err != nil {
		_ = os.Remove(cmd.TempPath)
		return reader.Book{}, err
	}

	go uc.process(context.WithoutCancel(ctx), book, cmd)
	return book, nil
}

func (uc *UploadBookUseCase) process(ctx context.Context, book reader.Book, cmd UploadBookCommand) {
	defer os.Remove(cmd.TempPath)

	if err := uc.convert(ctx, &book, cmd); err != nil {
		uc.logger.Error("reader convert failed", slog.String("book", book.ID), slog.String("err", err.Error()))
		book.Status = reader.StatusFailed
		book.Error = err.Error()
		if uErr := uc.books.Update(ctx, book); uErr != nil {
			uc.logger.Error("reader mark failed", slog.String("book", book.ID), slog.String("err", uErr.Error()))
		}
		return
	}

	book.Status = reader.StatusReady
	if err := uc.books.Update(ctx, book); err != nil {
		uc.logger.Error("reader mark ready", slog.String("book", book.ID), slog.String("err", err.Error()))
	}
}

func (uc *UploadBookUseCase) convert(ctx context.Context, book *reader.Book, cmd UploadBookCommand) error {
	workDir, err := os.MkdirTemp(uc.tempDir, "book-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	conv, err := uc.converter.Convert(ctx, cmd.TempPath, filepath.Join(workDir, "media"))
	if err != nil {
		return err
	}

	html := stripLeadingTitle(conv.HTML, book.Title)
	for _, m := range conv.Media {
		path, err := uc.putFile(ctx, book.ID+"/"+m.Ref, m.LocalPath, m.ContentType)
		if err != nil {
			return err
		}
		url := uc.urls.Build(path)
		html = strings.ReplaceAll(html, `"`+m.Ref+`"`, `"`+url+`"`)
		html = strings.ReplaceAll(html, `'`+m.Ref+`'`, `'`+url+`'`)
	}

	path, err := uc.putBytes(ctx, book.ID+"/content.html", []byte(html), "text/html; charset=utf-8")
	if err != nil {
		return err
	}
	book.ContentPath = path.String()
	book.TextLength = reader.TextLength(html)

	if cmd.Cover != nil && len(cmd.Cover.Data) > 0 {
		cover, err := uc.putBytes(ctx, book.ID+"/cover"+ext(cmd.Cover.Filename, ".jpg"), cmd.Cover.Data, cmd.Cover.ContentType)
		if err != nil {
			return err
		}
		book.CoverPath = cover.String()
	}

	if err := uc.analyze(ctx, book.ID, html); err != nil {
		uc.logger.Warn("reader lexicon analyze skipped", slog.String("book", book.ID), slog.String("err", err.Error()))
	}
	return nil
}

func (uc *UploadBookUseCase) analyze(ctx context.Context, mediaID, html string) error {
	if uc.analyzer == nil || uc.lex == nil {
		return nil
	}
	raw, err := uc.analyzer.Analyze(ctx, html)
	if err != nil {
		return fmt.Errorf("lexicon analyze: %w", err)
	}
	analysis := lexicon.MapAnalysis(raw)
	if err := uc.lex.SaveSentence(ctx, mediaID, analysis, lexicon.BuildSentenceSegments(analysis.Units)); err != nil {
		return fmt.Errorf("lexicon save: %w", err)
	}
	return nil
}

func (uc *UploadBookUseCase) putFile(ctx context.Context, key, localPath, contentType string) (media.Path, error) {
	// #nosec G304 -- localPath is a temp converter output generated by this process.
	f, err := os.Open(localPath)
	if err != nil {
		return media.Path{}, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return media.Path{}, err
	}
	return uc.put(ctx, key, f, stat.Size(), contentType)
}

func (uc *UploadBookUseCase) putBytes(ctx context.Context, key string, data []byte, contentType string) (media.Path, error) {
	return uc.put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
}

func (uc *UploadBookUseCase) put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (media.Path, error) {
	path, err := media.NewPath(uc.bucket + "/" + key)
	if err != nil {
		return media.Path{}, err
	}
	if err := uc.storage.Put(ctx, path, r, media.PutOptions{ContentType: contentType, Size: size}); err != nil {
		return media.Path{}, err
	}
	return path, nil
}

var (
	titleBlockRe = regexp.MustCompile(`(?is)^\s*<header[^>]*id="title-block-header"[^>]*>.*?</header>`)
	leadingH1Re  = regexp.MustCompile(`(?is)^\s*<h1[^>]*>(.*?)</h1>`)
	tagRe        = regexp.MustCompile(`<[^>]+>`)
)

// The reader UI always renders the book title above the content, so drop
// pandoc's generated title block and a leading H1 that repeats the title.
func stripLeadingTitle(html, title string) string {
	html = titleBlockRe.ReplaceAllString(html, "")
	for {
		m := leadingH1Re.FindStringSubmatch(html)
		if m == nil {
			break
		}
		text := strings.TrimSpace(gohtml.UnescapeString(tagRe.ReplaceAllString(m[1], "")))
		if !strings.EqualFold(text, strings.TrimSpace(title)) {
			break
		}
		html = html[len(m[0]):]
	}
	return strings.TrimSpace(html)
}

func bookTitle(cmd UploadBookCommand) string {
	if t := strings.TrimSpace(cmd.Title); t != "" {
		return t
	}
	base := filepath.Base(cmd.Filename)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func ext(filename, fallback string) string {
	if e := filepath.Ext(filename); e != "" {
		return strings.ToLower(e)
	}
	return fallback
}
