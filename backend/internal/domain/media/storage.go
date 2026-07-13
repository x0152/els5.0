package media

import (
	"context"
	"io"
	"time"
)

type PutOptions struct {
	ContentType string
	Size        int64
}

type Metadata struct {
	ContentType string
	Size        int64
}

type Storage interface {
	Put(ctx context.Context, path Path, r io.Reader, opts PutOptions) error

	Get(ctx context.Context, path Path) (io.ReadCloser, Metadata, error)

	Delete(ctx context.Context, path Path) error
}

type BucketEnsurer interface {
	EnsureBucket(ctx context.Context, bucket string) error
}

type URLSigner interface {
	SignedURL(ctx context.Context, path Path, ttl time.Duration) (string, error)
}
