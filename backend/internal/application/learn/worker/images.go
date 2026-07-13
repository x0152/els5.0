package worker

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/els/backend/internal/domain/illustration"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared/ports"
)

type Images struct {
	provider ports.ImageGenerator
	storage  media.Storage
	urls     media.PublicURL
	bucket   string
	logger   *slog.Logger
	running  sync.Map
	errs     sync.Map
}

func NewImages(provider ports.ImageGenerator, storage media.Storage, urls media.PublicURL, bucket string, logger *slog.Logger) *Images {
	if logger == nil {
		logger = slog.Default()
	}
	return &Images{provider: provider, storage: storage, urls: urls, bucket: bucket, logger: logger}
}

func (g *Images) available() bool {
	return g != nil && g.provider != nil && g.provider.IsAvailable() && g.storage != nil
}

func (g *Images) path(id string) (media.Path, error) {
	return media.NewPath(g.bucket + "/" + illustration.Filename(id))
}

func (g *Images) exists(ctx context.Context, path media.Path) bool {
	reader, _, err := g.storage.Get(ctx, path)
	if err != nil {
		return false
	}
	_ = reader.Close()
	return true
}

func (g *Images) Ensure(ctx context.Context, prompt, aspect string, trigger bool) illustration.Status {
	id := illustration.Key(prompt, aspect)
	path, err := g.path(id)
	if err != nil {
		return illustration.Status{ID: id, Status: illustration.StatusError, Error: "invalid prompt"}
	}

	if g.exists(ctx, path) {
		return illustration.Status{ID: id, Status: illustration.StatusReady, URL: g.urls.Build(path)}
	}
	if _, ok := g.running.Load(id); ok {
		return illustration.Status{ID: id, Status: illustration.StatusGenerating}
	}
	if msg, ok := g.errs.Load(id); ok && !trigger {
		return illustration.Status{ID: id, Status: illustration.StatusError, Error: msg.(string)}
	}
	if !trigger {
		return illustration.Status{ID: id, Status: illustration.StatusPending}
	}

	if !g.available() {
		return illustration.Status{ID: id, Status: illustration.StatusError, Error: "image generation is not available"}
	}
	if _, loaded := g.running.LoadOrStore(id, struct{}{}); loaded {
		return illustration.Status{ID: id, Status: illustration.StatusGenerating}
	}
	g.errs.Delete(id)
	go g.generate(context.WithoutCancel(ctx), id, path, prompt, aspect)
	return illustration.Status{ID: id, Status: illustration.StatusGenerating}
}

func toAspect(aspect string) ports.ImageAspect {
	switch aspect {
	case string(ports.ImageAspectLandscape):
		return ports.ImageAspectLandscape
	case string(ports.ImageAspectPortrait):
		return ports.ImageAspectPortrait
	default:
		return ports.ImageAspectSquare
	}
}

func (g *Images) generate(ctx context.Context, id string, path media.Path, prompt, aspect string) {
	defer g.running.Delete(id)
	defer func() {
		if r := recover(); r != nil {
			g.errs.Store(id, "internal error")
			g.logger.Error("illustration panic", slog.String("id", id), slog.Any("panic", r))
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	data, err := g.provider.GenerateImageBytes(ctx, illustration.StyledPrompt(prompt), &ports.ImageOptions{Aspect: toAspect(aspect)})
	if err != nil {
		g.errs.Store(id, truncate(err.Error()))
		return
	}
	if err := g.storage.Put(ctx, path, bytes.NewReader(data), media.PutOptions{ContentType: "image/png", Size: int64(len(data))}); err != nil {
		g.errs.Store(id, "failed to store image")
		g.logger.Warn("illustration store failed", slog.String("id", id), slog.String("err", err.Error()))
	}
}

func truncate(msg string) string {
	msg = strings.ReplaceAll(strings.TrimSpace(msg), "\n", " ")
	if len(msg) > 200 {
		return msg[:200] + "…"
	}
	return msg
}
