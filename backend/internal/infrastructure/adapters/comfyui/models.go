package comfyui

import (
	"context"
	"net/http"
	"time"

	"github.com/els/backend/internal/domain/shared/ports"
)

type ModelLister struct {
	httpClient *http.Client
}

func NewModelLister() *ModelLister {
	return &ModelLister{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (l *ModelLister) ListModels(ctx context.Context, cfg ports.AIProviderConfig) ([]string, error) {
	return ListCheckpoints(ctx, l.httpClient, cfg)
}

var _ ports.ModelLister = (*ModelLister)(nil)
