package pandoc

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/els/backend/internal/domain/shared/ports"
)

type Converter struct{}

func New() *Converter { return &Converter{} }

var (
	bodyOpenRe  = regexp.MustCompile(`(?is)^.*?<body[^>]*>`)
	bodyCloseRe = regexp.MustCompile(`(?is)</body>.*$`)
)

func (c *Converter) Convert(ctx context.Context, srcPath, mediaDir string) (ports.BookConversion, error) {
	workDir := filepath.Dir(mediaDir)
	mediaName := filepath.Base(mediaDir)
	outPath := filepath.Join(workDir, "output.html")

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "pandoc", srcPath, "-o", outPath, "--standalone", "--extract-media="+mediaName) // #nosec G204 -- fixed binary with argv, no shell expansion.
	cmd.Dir = workDir
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}
		return ports.BookConversion{}, fmt.Errorf("pandoc convert: %s", detail)
	}

	raw, err := os.ReadFile(outPath) // #nosec G304 -- outPath is generated inside the temp workDir.
	if err != nil {
		return ports.BookConversion{}, err
	}
	html := bodyCloseRe.ReplaceAllString(bodyOpenRe.ReplaceAllString(string(raw), ""), "")

	media := []ports.BookMedia{}
	_ = filepath.WalkDir(mediaDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(workDir, path)
		if err != nil {
			return nil
		}
		media = append(media, ports.BookMedia{
			Ref:         filepath.ToSlash(rel),
			LocalPath:   path,
			ContentType: mime.TypeByExtension(filepath.Ext(path)),
		})
		return nil
	})

	return ports.BookConversion{HTML: strings.TrimSpace(html), Media: media}, nil
}
