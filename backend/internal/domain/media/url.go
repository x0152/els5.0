package media

import (
	"net/url"
	"strings"
)

type PublicURL struct {
	base string
}

func NewPublicURL(base string) PublicURL {
	return PublicURL{base: strings.TrimRight(strings.TrimSpace(base), "/")}
}

func (b PublicURL) Base() string { return b.base }

func (b PublicURL) Build(p Path) string {
	if p.IsZero() {
		return ""
	}
	return b.base + "/" + p.Bucket() + "/" + escapeKey(p.Key())
}

func (b PublicURL) BuildFromRaw(raw string) string {
	p, err := NewPath(raw)
	if err != nil {
		return ""
	}
	return b.Build(p)
}

func (b PublicURL) ParsePath(publicURL string) (Path, bool) {
	if publicURL == "" {
		return Path{}, false
	}
	prefix := b.base + "/"
	if !strings.HasPrefix(publicURL, prefix) {
		return Path{}, false
	}
	rel := strings.TrimPrefix(publicURL, prefix)
	unescaped, err := url.PathUnescape(rel)
	if err != nil {
		return Path{}, false
	}
	p, err := NewPath(unescaped)
	if err != nil {
		return Path{}, false
	}
	return p, true
}

func escapeKey(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
