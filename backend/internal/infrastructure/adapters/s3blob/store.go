package s3blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/els/backend/internal/domain/media"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

type Store struct {
	client *minio.Client
	region string
}

func New(cfg Config) (*Store, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("s3blob: endpoint must not be empty")
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, errors.New("s3blob: access/secret key must not be empty")
	}
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("s3blob: init client: %w", err)
	}
	return &Store{client: cli, region: cfg.Region}, nil
}

func (s *Store) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("s3blob: check bucket %q: %w", bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: s.region}); err != nil {
		exists, eerr := s.client.BucketExists(ctx, bucket)
		if eerr == nil && exists {
			return nil
		}
		return fmt.Errorf("s3blob: create bucket %q: %w", bucket, err)
	}
	return nil
}

func (s *Store) Put(ctx context.Context, path media.Path, r io.Reader, opts media.PutOptions) error {
	if path.IsZero() {
		return errors.New("s3blob: empty path")
	}
	if opts.Size <= 0 {
		return errors.New("s3blob: size must be > 0")
	}
	if _, err := s.client.PutObject(ctx, path.Bucket(), path.Key(), r, opts.Size, minio.PutObjectOptions{
		ContentType: opts.ContentType,
	}); err != nil {
		return fmt.Errorf("s3blob: put %s: %w", path, err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, path media.Path) (io.ReadCloser, media.Metadata, error) {
	if path.IsZero() {
		return nil, media.Metadata{}, errors.New("s3blob: empty path")
	}
	obj, err := s.client.GetObject(ctx, path.Bucket(), path.Key(), minio.GetObjectOptions{})
	if err != nil {
		return nil, media.Metadata{}, fmt.Errorf("s3blob: get %s: %w", path, err)
	}
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		if isNotFound(err) {
			return nil, media.Metadata{}, media.ErrNotFound
		}
		return nil, media.Metadata{}, fmt.Errorf("s3blob: stat %s: %w", path, err)
	}
	return obj, media.Metadata{
		ContentType: info.ContentType,
		Size:        info.Size,
	}, nil
}

func (s *Store) SignedURL(ctx context.Context, path media.Path, ttl time.Duration) (string, error) {
	if path.IsZero() {
		return "", errors.New("s3blob: empty path")
	}
	u, err := s.client.PresignedGetObject(ctx, path.Bucket(), path.Key(), ttl, url.Values{})
	if err != nil {
		return "", fmt.Errorf("s3blob: presign %s: %w", path, err)
	}
	return u.String(), nil
}

func (s *Store) Delete(ctx context.Context, path media.Path) error {
	if path.IsZero() {
		return nil
	}
	if err := s.client.RemoveObject(ctx, path.Bucket(), path.Key(), minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("s3blob: delete %s: %w", path, err)
	}
	return nil
}

func isNotFound(err error) bool {
	var er minio.ErrorResponse
	if errors.As(err, &er) {
		return er.StatusCode == 404 || er.Code == "NoSuchKey" || er.Code == "NoSuchBucket"
	}
	return false
}

var _ media.Storage = (*Store)(nil)
