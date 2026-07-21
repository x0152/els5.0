package imagegen

import (
	"context"
	"time"

	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/bothub"
	"github.com/els/backend/internal/infrastructure/adapters/comfyui"
)

// Router dispatches image generation to the engine selected in settings:
// an OpenAI-compatible API (bothub) or a ComfyUI server.
type Router struct {
	resolver ports.AIProviderResolver
	openai   ports.ImageGenerator
	comfy    ports.ImageGenerator
}

func NewWithResolver(baseURL, apiKey, model string, timeout time.Duration, resolver ports.AIProviderResolver) *Router {
	return &Router{
		resolver: resolver,
		openai:   bothub.NewWithResolver(baseURL, apiKey, model, timeout, resolver),
		comfy:    comfyui.NewWithResolver(timeout, resolver),
	}
}

func (r *Router) pick(ctx context.Context) ports.ImageGenerator {
	if r.resolver != nil && r.resolver.Resolve(ctx).Kind == string(settings.KindComfyUI) {
		return r.comfy
	}
	return r.openai
}

func (r *Router) IsAvailable() bool {
	return r.pick(context.Background()).IsAvailable()
}

func (r *Router) GenerateImageBytes(ctx context.Context, prompt string, opts *ports.ImageOptions) ([]byte, error) {
	return r.pick(ctx).GenerateImageBytes(ctx, prompt, opts)
}

var _ ports.ImageGenerator = (*Router)(nil)
