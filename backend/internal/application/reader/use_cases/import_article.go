package usecases

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/reader"
	"github.com/els/backend/internal/domain/shared"
)

type ImportArticleUseCase struct {
	upload  *UploadBookUseCase
	tempDir string
}

func NewImportArticleUseCase(upload *UploadBookUseCase, tempDir string) *ImportArticleUseCase {
	return &ImportArticleUseCase{upload: upload, tempDir: tempDir}
}

type ImportArticleCommand struct {
	URL        string
	GroupTitle string
}

func (uc *ImportArticleUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd ImportArticleCommand) (reader.Book, error) {
	if actor == nil {
		return reader.Book{}, shared.ErrUnauthorized
	}
	url := strings.TrimSpace(cmd.URL)
	if url == "" {
		return reader.Book{}, shared.Validation(fmt.Errorf("url: must not be empty"))
	}

	article, err := readability.FromURL(url, 30*time.Second, func(r *http.Request) {
		r.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ELSReader/1.0)")
	})
	if err != nil || article.Node == nil {
		return reader.Book{}, shared.Validation(fmt.Errorf("url: could not extract article from this page"))
	}

	var html bytes.Buffer
	if err := article.RenderHTML(&html); err != nil {
		return reader.Book{}, err
	}

	f, err := os.CreateTemp(uc.tempDir, "article-*.html")
	if err != nil {
		return reader.Book{}, err
	}
	if _, err := f.Write(html.Bytes()); err != nil {
		f.Close()
		os.Remove(f.Name())
		return reader.Book{}, err
	}
	f.Close()

	return uc.upload.Execute(ctx, actor, UploadBookCommand{
		Title:      article.Title(),
		Author:     article.Byline(),
		Filename:   "article.html",
		TempPath:   f.Name(),
		Kind:       reader.KindArticle,
		GroupTitle: cmd.GroupTitle,
		Cover:      fetchCover(ctx, article.ImageURL()),
	})
}

func fetchCover(ctx context.Context, rawURL string) *UploadAsset {
	if strings.TrimSpace(rawURL) == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ELSReader/1.0)")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/") {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil || len(data) == 0 {
		return nil
	}
	name := "cover"
	if u, err := url.Parse(rawURL); err == nil {
		if b := path.Base(u.Path); b != "" && b != "/" && b != "." {
			name = b
		}
	}
	return &UploadAsset{Data: data, ContentType: resp.Header.Get("Content-Type"), Filename: name}
}
