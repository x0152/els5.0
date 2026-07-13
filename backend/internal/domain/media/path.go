package media

import (
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/shared"
)

type Path struct {
	raw string
}

func NewPath(raw string) (Path, error) {
	trimmed := strings.Trim(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return Path{}, fmt.Errorf("%w: media.path: must not be empty", shared.ErrValidation)
	}
	bucket, key, ok := strings.Cut(trimmed, "/")
	if !ok || bucket == "" || key == "" {
		return Path{}, fmt.Errorf("%w: media.path: must be %q (first segment is bucket)", shared.ErrValidation, "<bucket>/<key>")
	}
	if strings.Contains(bucket, "..") || strings.Contains(key, "..") {
		return Path{}, fmt.Errorf("%w: media.path: must not contain '..'", shared.ErrValidation)
	}
	return Path{raw: trimmed}, nil
}

func (p Path) String() string { return p.raw }
func (p Path) IsZero() bool   { return p.raw == "" }

func (p Path) Bucket() string {
	bucket, _, _ := strings.Cut(p.raw, "/")
	return bucket
}

func (p Path) Key() string {
	_, key, _ := strings.Cut(p.raw, "/")
	return key
}
