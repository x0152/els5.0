package ports

import "context"

type BookMedia struct {
	Ref         string
	LocalPath   string
	ContentType string
}

type BookConversion struct {
	HTML  string
	Media []BookMedia
}

type BookConverter interface {
	Convert(ctx context.Context, srcPath, mediaDir string) (BookConversion, error)
}
