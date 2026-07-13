package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared/ports"
)

type VisionClient struct {
	cfg      ports.AIProviderConfig
	resolver ports.AIProviderResolver
	http     *http.Client
}

func NewVisionClient(baseURL, apiKey, model string, timeout time.Duration, resolver ports.AIProviderResolver) *VisionClient {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &VisionClient{
		cfg: ports.AIProviderConfig{
			BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
			APIKey:  strings.TrimSpace(apiKey),
			Model:   strings.TrimSpace(model),
		},
		resolver: resolver,
		http:     &http.Client{Timeout: timeout},
	}
}

func (c *VisionClient) resolve(ctx context.Context) ports.AIProviderConfig {
	if c.resolver != nil {
		if r := c.resolver.Resolve(ctx); !r.IsEmpty() {
			return ports.AIProviderConfig{
				BaseURL: strings.TrimRight(strings.TrimSpace(r.BaseURL), "/"),
				APIKey:  strings.TrimSpace(r.APIKey),
				Model:   strings.TrimSpace(r.Model),
			}
		}
	}
	return c.cfg
}

func (c *VisionClient) Describe(ctx context.Context, image []byte, mime, question string) (string, error) {
	cfg := c.resolve(ctx)
	if cfg.BaseURL == "" || cfg.Model == "" {
		return "", fmt.Errorf("llm: vision provider not configured")
	}
	if mime == "" {
		mime = "image/jpeg"
	}
	dataURL := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(image)
	body, _ := json.Marshal(map[string]any{
		"model": cfg.Model,
		"messages": []map[string]any{{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": question},
				{"type": "image_url", "image_url": map[string]any{"url": dataURL}},
			},
		}},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm vision status %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("llm vision empty response")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
