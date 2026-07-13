package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/agent"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared/ports"
)

type ImagePlugin struct{ tools []agent.Tool }

func NewImagePlugin(gen ports.ImageGenerator, storage media.Storage, urls media.PublicURL, bucket string) *ImagePlugin {
	if gen == nil || storage == nil {
		return &ImagePlugin{}
	}
	return &ImagePlugin{tools: []agent.Tool{generateImage(gen, storage, urls, bucket)}}
}

func (p *ImagePlugin) Tools(_ agent.RunContext) []agent.Tool { return p.tools }

func generateImage(gen ports.ImageGenerator, storage media.Storage, urls media.PublicURL, bucket string) agent.Tool {
	return agent.Tool{
		Name:        "generate_image",
		Description: "Generates an illustration from a description (e.g. for a word, phrase, or scene) and returns ready markdown image markup. Always insert the returned markup into your reply — then the image will show in the chat.",
		Icon:        "image",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prompt":  map[string]any{"type": "string", "description": "What to depict in the image, detailed description in English."},
				"caption": map[string]any{"type": "string", "description": "Short caption/alt, e.g. the word itself."},
			},
			"required": []string{"prompt"},
		},
		Label: func(args string) string {
			var a struct {
				Caption string `json:"caption"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if a.Caption != "" {
				return fmt.Sprintf("Drawing \"%s\"", a.Caption)
			}
			return "Generating an image"
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			if !gen.IsAvailable() {
				return "Image generation is unavailable.", nil
			}
			var a struct {
				Prompt  string `json:"prompt"`
				Caption string `json:"caption"`
			}
			if err := json.Unmarshal([]byte(args), &a); err != nil {
				return "", fmt.Errorf("invalid arguments: %w", err)
			}
			if strings.TrimSpace(a.Prompt) == "" {
				return "Provide a prompt for the image.", nil
			}
			data, err := gen.GenerateImageBytes(ctx, a.Prompt, &ports.ImageOptions{Aspect: ports.ImageAspectSquare})
			if err != nil {
				return "", err
			}
			path, err := media.NewPath(bucket + "/" + uuid.NewString() + ".png")
			if err != nil {
				return "", err
			}
			if err := storage.Put(ctx, path, bytes.NewReader(data), media.PutOptions{ContentType: "image/png", Size: int64(len(data))}); err != nil {
				return "", err
			}
			caption := strings.TrimSpace(a.Caption)
			if caption == "" {
				caption = "image"
			}
			return fmt.Sprintf("![%s](%s)", caption, urls.Build(path)), nil
		},
	}
}
