package filmvision

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/media"
)

type FrameExtractor interface {
	ExtractFrameJPEG(ctx context.Context, src string, atSeconds float64) ([]byte, error)
}

type VisionDescriber interface {
	Describe(ctx context.Context, image []byte, mime, question string) (string, error)
}

type Service struct {
	films   films.Repository
	storage media.Storage
	signer  media.URLSigner
	frames  FrameExtractor
	vision  VisionDescriber
}

func NewService(repo films.Repository, storage media.Storage, frames FrameExtractor, vision VisionDescriber) *Service {
	s := &Service{films: repo, storage: storage, frames: frames, vision: vision}
	if signer, ok := storage.(media.URLSigner); ok {
		s.signer = signer
	}
	return s
}

func (s *Service) ReadFrame(ctx context.Context, filmID string, atMs int, question string) (string, error) {
	film, err := s.films.Get(ctx, filmID)
	if err != nil {
		return "", err
	}
	if len(film.AudioVariants) == 0 {
		return "The film has no video track for frame extraction.", nil
	}
	path, err := media.NewPath(film.AudioVariants[0].Path)
	if err != nil {
		return "Film video is unavailable.", nil
	}
	if atMs < 0 {
		atMs = 0
	}

	src, cleanup, err := s.source(ctx, path)
	if err != nil {
		return "", err
	}
	defer cleanup()

	frame, err := s.frames.ExtractFrameJPEG(ctx, src, float64(atMs)/1000)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(question) == "" {
		question = "Briefly describe what is happening in this film frame: people, actions, setting, on-screen text."
	}
	return s.vision.Describe(ctx, frame, "image/jpeg", question)
}

func (s *Service) source(ctx context.Context, path media.Path) (string, func(), error) {
	if s.signer != nil {
		url, err := s.signer.SignedURL(ctx, path, 10*time.Minute)
		if err == nil {
			return url, func() {}, nil
		}
	}
	rc, _, err := s.storage.Get(ctx, path)
	if err != nil {
		return "", func() {}, err
	}
	defer rc.Close()
	tmp, err := os.CreateTemp("", "frame-*.mp4")
	if err != nil {
		return "", func() {}, err
	}
	if _, err := io.Copy(tmp, rc); err != nil {
		tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", func() {}, fmt.Errorf("download video: %w", err)
	}
	tmp.Close()
	return tmp.Name(), func() { _ = os.Remove(tmp.Name()) }, nil
}
