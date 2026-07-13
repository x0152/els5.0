package media

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/els/backend/internal/domain/shared"
)

type Kind string

const (
	KindImage    Kind = "image"
	KindDocument Kind = "document"
)

const (
	defaultImageMaxBytes    int64 = 5 * 1024 * 1024
	defaultDocumentMaxBytes int64 = 25 * 1024 * 1024
)

type Format struct {
	Mime          string
	Extension     string
	AltExtensions []string
}

type UploadPolicy struct {
	Kind    Kind
	MaxSize int64
	Formats []Format
}

func (p UploadPolicy) Validate(size int64, mime, filename string) (string, error) {
	if size <= 0 {
		return "", fmt.Errorf("%w: file.size: must be > 0", shared.ErrValidation)
	}
	if p.MaxSize > 0 && size > p.MaxSize {
		return "", fmt.Errorf("%w: file.size: max %d bytes", shared.ErrValidation, p.MaxSize)
	}
	mime = strings.ToLower(strings.TrimSpace(mime))
	for i := range p.Formats {
		if p.Formats[i].Mime == mime {
			return p.pickExtension(p.Formats[i], filename), nil
		}
	}
	return "", fmt.Errorf("%w: file.content_type: %q is not allowed for %s; allowed: %s",
		shared.ErrValidation, mime, p.Kind, strings.Join(p.AllowedMimes(), ", "))
}

func (p UploadPolicy) AllowedMimes() []string {
	out := make([]string, 0, len(p.Formats))
	for _, f := range p.Formats {
		out = append(out, f.Mime)
	}
	return out
}

func (p UploadPolicy) pickExtension(f Format, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" && (ext == f.Extension || slices.Contains(f.AltExtensions, ext)) {
		return ext
	}
	return f.Extension
}

func ImagePolicy(maxSize int64) UploadPolicy {
	if maxSize <= 0 {
		maxSize = defaultImageMaxBytes
	}
	return UploadPolicy{
		Kind:    KindImage,
		MaxSize: maxSize,
		Formats: []Format{
			{Mime: "image/png", Extension: ".png"},
			{Mime: "image/jpeg", Extension: ".jpg", AltExtensions: []string{".jpeg"}},
			{Mime: "image/webp", Extension: ".webp"},
			{Mime: "image/gif", Extension: ".gif"},
		},
	}
}

func DocumentPolicy(maxSize int64) UploadPolicy {
	if maxSize <= 0 {
		maxSize = defaultDocumentMaxBytes
	}
	return UploadPolicy{
		Kind:    KindDocument,
		MaxSize: maxSize,
		Formats: []Format{
			{Mime: "application/pdf", Extension: ".pdf"},
			{Mime: "application/zip", Extension: ".zip", AltExtensions: []string{".docx", ".xlsx", ".pptx"}},
		},
	}
}
