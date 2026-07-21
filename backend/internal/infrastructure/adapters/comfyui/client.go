package comfyui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared/ports"
)

// Client generates images through a ComfyUI server using a standard
// txt2img workflow (checkpoint -> prompts -> sampler -> save).
type Client struct {
	httpClient *http.Client
	resolver   ports.AIProviderResolver
}

func NewWithResolver(timeout time.Duration, resolver ports.AIProviderResolver) *Client {
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	return &Client{httpClient: &http.Client{Timeout: timeout}, resolver: resolver}
}

func (c *Client) config(ctx context.Context) ports.AIProviderConfig {
	if c.resolver == nil {
		return ports.AIProviderConfig{}
	}
	return c.resolver.Resolve(ctx)
}

func (c *Client) IsAvailable() bool {
	cfg := c.config(context.Background())
	return strings.TrimSpace(cfg.BaseURL) != "" && strings.TrimSpace(cfg.Model) != ""
}

func param(cfg ports.AIProviderConfig, key, def string) string {
	if v := strings.TrimSpace(cfg.Params[key]); v != "" {
		return v
	}
	return def
}

func paramInt(cfg ports.AIProviderConfig, key string, def int) int {
	if n, err := strconv.Atoi(param(cfg, key, "")); err == nil && n > 0 {
		return n
	}
	return def
}

func paramFloat(cfg ports.AIProviderConfig, key string, def float64) float64 {
	if f, err := strconv.ParseFloat(param(cfg, key, ""), 64); err == nil && f > 0 {
		return f
	}
	return def
}

func dimensions(cfg ports.AIProviderConfig, aspect ports.ImageAspect) (int, int) {
	w := paramInt(cfg, "width", 1024)
	h := paramInt(cfg, "height", 1024)
	switch aspect {
	case ports.ImageAspectLandscape:
		return maxInt(w, h) * 4 / 3, minInt(w, h)
	case ports.ImageAspectWide:
		return maxInt(w, h) * 7 / 4, minInt(w, h)
	case ports.ImageAspectPortrait:
		return minInt(w, h), maxInt(w, h) * 4 / 3
	default:
		return w, h
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (c *Client) GenerateImageBytes(ctx context.Context, prompt string, opts *ports.ImageOptions) ([]byte, error) {
	cfg := c.config(ctx)
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" || strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("comfyui is not configured")
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	aspect := ports.ImageAspectSquare
	if opts != nil && opts.Aspect != "" {
		aspect = opts.Aspect
	}
	width, height := dimensions(cfg, aspect)

	workflow := map[string]any{
		"ckpt": map[string]any{
			"class_type": "CheckpointLoaderSimple",
			"inputs":     map[string]any{"ckpt_name": cfg.Model},
		},
		"pos": map[string]any{
			"class_type": "CLIPTextEncode",
			"inputs":     map[string]any{"clip": []any{"ckpt", 1}, "text": prompt},
		},
		"neg": map[string]any{
			"class_type": "CLIPTextEncode",
			"inputs":     map[string]any{"clip": []any{"ckpt", 1}, "text": param(cfg, "negative_prompt", "text, watermark, low quality")},
		},
		"latent": map[string]any{
			"class_type": "EmptyLatentImage",
			"inputs":     map[string]any{"width": width, "height": height, "batch_size": 1},
		},
		"sampler": map[string]any{
			"class_type": "KSampler",
			"inputs": map[string]any{
				"model":        []any{"ckpt", 0},
				"positive":     []any{"pos", 0},
				"negative":     []any{"neg", 0},
				"latent_image": []any{"latent", 0},
				"seed":         rand.Int63(),
				"steps":        paramInt(cfg, "steps", 25),
				"cfg":          paramFloat(cfg, "cfg", 6.5),
				"sampler_name": param(cfg, "sampler", "euler"),
				"scheduler":    param(cfg, "scheduler", "normal"),
				"denoise":      1.0,
			},
		},
		"decode": map[string]any{
			"class_type": "VAEDecode",
			"inputs":     map[string]any{"samples": []any{"sampler", 0}, "vae": []any{"ckpt", 2}},
		},
		"save": map[string]any{
			"class_type": "SaveImage",
			"inputs":     map[string]any{"images": []any{"decode", 0}, "filename_prefix": "els"},
		},
	}

	promptID, err := c.queue(ctx, baseURL, cfg.APIKey, workflow)
	if err != nil {
		return nil, err
	}
	return c.waitForImage(ctx, baseURL, cfg.APIKey, promptID)
}

func (c *Client) queue(ctx context.Context, baseURL, apiKey string, workflow map[string]any) (string, error) {
	body, _ := json.Marshal(map[string]any{"prompt": workflow})
	respBody, err := c.do(ctx, http.MethodPost, baseURL+"/prompt", apiKey, body)
	if err != nil {
		return "", err
	}
	var out struct {
		PromptID string `json:"prompt_id"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil || out.PromptID == "" {
		return "", fmt.Errorf("comfyui queue: unexpected response: %s", strings.TrimSpace(string(respBody)))
	}
	return out.PromptID, nil
}

func (c *Client) waitForImage(ctx context.Context, baseURL, apiKey, promptID string) ([]byte, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}

		respBody, err := c.do(ctx, http.MethodGet, baseURL+"/history/"+promptID, apiKey, nil)
		if err != nil {
			return nil, err
		}
		var history map[string]struct {
			Status struct {
				Completed bool   `json:"completed"`
				StatusStr string `json:"status_str"`
			} `json:"status"`
			Outputs map[string]struct {
				Images []struct {
					Filename  string `json:"filename"`
					Subfolder string `json:"subfolder"`
					Type      string `json:"type"`
				} `json:"images"`
			} `json:"outputs"`
		}
		if err := json.Unmarshal(respBody, &history); err != nil {
			return nil, fmt.Errorf("comfyui history: %w", err)
		}
		entry, ok := history[promptID]
		if !ok {
			continue
		}
		if entry.Status.StatusStr == "error" {
			return nil, fmt.Errorf("comfyui generation failed")
		}
		for _, out := range entry.Outputs {
			for _, img := range out.Images {
				if img.Type != "output" {
					continue
				}
				q := url.Values{"filename": {img.Filename}, "subfolder": {img.Subfolder}, "type": {img.Type}}
				return c.do(ctx, http.MethodGet, baseURL+"/view?"+q.Encode(), apiKey, nil)
			}
		}
		if entry.Status.Completed {
			return nil, fmt.Errorf("comfyui: no output image")
		}
	}
}

func (c *Client) do(ctx context.Context, method, rawURL, apiKey string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("comfyui request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("comfyui HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return respBody, nil
}

// ListCheckpoints returns checkpoint names available on the ComfyUI server.
func ListCheckpoints(ctx context.Context, httpClient *http.Client, cfg ports.AIProviderConfig) ([]string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("comfyui: base url not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/object_info/CheckpointLoaderSimple", nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("comfyui request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("comfyui HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var out struct {
		Node struct {
			Input struct {
				Required struct {
					CkptName []json.RawMessage `json:"ckpt_name"`
				} `json:"required"`
			} `json:"input"`
		} `json:"CheckpointLoaderSimple"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("comfyui object_info: %w", err)
	}
	if len(out.Node.Input.Required.CkptName) == 0 {
		return nil, nil
	}
	var names []string
	if err := json.Unmarshal(out.Node.Input.Required.CkptName[0], &names); err != nil {
		return nil, fmt.Errorf("comfyui checkpoints: %w", err)
	}
	return names, nil
}

var _ ports.ImageGenerator = (*Client)(nil)
