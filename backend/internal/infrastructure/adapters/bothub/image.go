package bothub

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared/ports"
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	available  bool
	resolver   ports.AIProviderResolver
}

var imageSizePattern = regexp.MustCompile(`^\d{2,5}x\d{2,5}$`)

func New(baseURL, apiKey, model string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 180 * time.Second
	}
	c := &Client{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		httpClient: &http.Client{Timeout: timeout},
	}
	c.available = c.baseURL != "" && c.apiKey != "" && c.model != ""
	return c
}

func NewWithResolver(baseURL, apiKey, model string, timeout time.Duration, resolver ports.AIProviderResolver) *Client {
	c := New(baseURL, apiKey, model, timeout)
	c.resolver = resolver
	return c
}

func (c *Client) effective(ctx context.Context) *Client {
	if c.resolver == nil {
		return c
	}
	r := c.resolver.Resolve(ctx)
	if r.IsEmpty() {
		return c
	}
	ec := *c
	ec.baseURL = strings.TrimRight(strings.TrimSpace(r.BaseURL), "/")
	ec.apiKey = strings.TrimSpace(r.APIKey)
	ec.model = strings.TrimSpace(r.Model)
	ec.available = ec.baseURL != "" && ec.apiKey != "" && ec.model != ""
	return &ec
}

func (c *Client) IsAvailable() bool { return c.effective(context.Background()).available }

func (c *Client) GenerateImageBytes(ctx context.Context, prompt string, opts *ports.ImageOptions) ([]byte, error) {
	c = c.effective(ctx)
	if !c.available {
		return nil, fmt.Errorf("bothub is unavailable")
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	aspect := ports.ImageAspectSquare
	if opts != nil && opts.Aspect != "" {
		aspect = opts.Aspect
	}
	size := ""
	references := []ports.ImageReference{}
	if opts != nil {
		size = normalizeSize(opts.Size)
		references = opts.References
	}

	var (
		imageData []byte
		lastErr   error
		backoff   = 3 * time.Second
	)
	for attempt := 0; attempt < 4; attempt++ {
		imageData, lastErr = c.callAPI(ctx, prompt, aspect, size, references)
		if lastErr != nil && size != "" && unsupportedImageSizeErr(lastErr) {
			size = ""
			imageData, lastErr = c.callAPI(ctx, prompt, aspect, "", references)
		}
		if lastErr == nil {
			return imageData, nil
		}

		errText := lastErr.Error()
		retryable := strings.Contains(errText, "429") || strings.Contains(errText, "HTTP 5") ||
			strings.Contains(errText, "HTTP 403") || strings.Contains(errText, "INVALID_MODEL_PRICING")
		if !retryable {
			return nil, lastErr
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return nil, lastErr
}

func (c *Client) callAPI(ctx context.Context, prompt string, aspect ports.ImageAspect, size string, references []ports.ImageReference) ([]byte, error) {
	if strings.EqualFold(c.model, "dall-e-3") {
		if len(references) == 0 {
			return c.callImagesEndpoint(ctx, prompt, aspect, size)
		}
		return c.callChatCompletions(ctx, prompt, aspect, size, references)
	}
	if len(references) == 0 {
		if data, err := c.callImagesEndpoint(ctx, prompt, aspect, size); err == nil {
			return data, nil
		}
		return c.callChatCompletions(ctx, prompt, aspect, size, references)
	}
	data, err := c.callChatCompletions(ctx, prompt, aspect, size, references)
	if err != nil && strings.Contains(err.Error(), "INVALID_MODEL_PRICING") {
		// Some models are billable only on /images/generations; drop the
		// reference images and generate without them instead of failing.
		return c.callImagesEndpoint(ctx, prompt, aspect, size)
	}
	return data, err
}

func sizeForAspect(aspect ports.ImageAspect) string {
	switch aspect {
	case ports.ImageAspectLandscape, ports.ImageAspectWide:
		return "1792x1024"
	case ports.ImageAspectPortrait:
		return "1024x1792"
	default:
		return "1024x1024"
	}
}

func (c *Client) callImagesEndpoint(ctx context.Context, prompt string, aspect ports.ImageAspect, size string) ([]byte, error) {
	reqSize := sizeForAspect(aspect)
	if size != "" {
		reqSize = size
	}
	body, _ := json.Marshal(map[string]any{
		"model":  c.model,
		"prompt": prompt,
		"n":      1,
		"size":   reqSize,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/images/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bothub images request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bothub images HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse images response: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no image in response")
	}

	item := result.Data[0]
	if strings.TrimSpace(item.B64JSON) != "" {
		return base64.StdEncoding.DecodeString(item.B64JSON)
	}
	if strings.TrimSpace(item.URL) != "" {
		return c.downloadURL(ctx, item.URL)
	}
	return nil, fmt.Errorf("no url or b64_json in response")
}

func (c *Client) callChatCompletions(ctx context.Context, prompt string, aspect ports.ImageAspect, size string, references []ports.ImageReference) ([]byte, error) {
	aspectHint := ""
	switch aspect {
	case ports.ImageAspectLandscape:
		aspectHint = " Use landscape orientation (wide, roughly 16:9 aspect ratio)."
	case ports.ImageAspectWide:
		aspectHint = " Use an ultra-wide cinematic banner orientation (roughly 21:9 aspect ratio), with the key subject centered and empty breathing room on the left and right edges so nothing important is cropped."
	case ports.ImageAspectPortrait:
		aspectHint = " Use portrait orientation (tall, roughly 9:16 aspect ratio)."
	}
	sizeHint := ""
	if size != "" {
		sizeHint = " Use exact image size " + size + "."
	}
	text := "Generate an image: " + prompt + aspectHint + sizeHint
	body, _ := json.Marshal(map[string]any{
		"model": c.model,
		"messages": []map[string]any{
			{"role": "user", "content": chatContentWithReferences(text, references)},
		},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bothub chat request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bothub chat HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Images []struct {
					ImageURL struct {
						URL string `json:"url"`
					} `json:"image_url"`
				} `json:"images"`
				InlineData *struct {
					Data string `json:"data"`
				} `json:"inline_data"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse chat response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	message := result.Choices[0].Message
	if message.InlineData != nil && strings.TrimSpace(message.InlineData.Data) != "" {
		return base64.StdEncoding.DecodeString(message.InlineData.Data)
	}

	if len(message.Images) > 0 {
		raw := strings.TrimSpace(message.Images[0].ImageURL.URL)
		if strings.HasPrefix(raw, "data:") {
			parts := strings.SplitN(raw, ",", 2)
			if len(parts) == 2 {
				return base64.StdEncoding.DecodeString(parts[1])
			}
		}
		return c.downloadURL(ctx, raw)
	}

	return nil, fmt.Errorf("no image data in chat response")
}

func chatContentWithReferences(text string, references []ports.ImageReference) any {
	parts := []map[string]any{
		{
			"type": "text",
			"text": text,
		},
	}
	for _, reference := range references {
		if len(reference.Data) == 0 {
			continue
		}
		contentType := strings.TrimSpace(strings.ToLower(reference.ContentType))
		if contentType == "" || !strings.HasPrefix(contentType, "image/") {
			contentType = "image/png"
		}
		dataURL := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(reference.Data)
		parts = append(parts, map[string]any{
			"type": "image_url",
			"image_url": map[string]any{
				"url": dataURL,
			},
		})
	}
	if len(parts) == 1 {
		return text
	}
	return parts
}

func (c *Client) downloadURL(ctx context.Context, rawURL string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("download HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	return io.ReadAll(resp.Body)
}

func normalizeSize(raw string) string {
	size := strings.ToLower(strings.TrimSpace(raw))
	if !imageSizePattern.MatchString(size) {
		return ""
	}
	return size
}

func unsupportedImageSizeErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid size") ||
		strings.Contains(msg, "unsupported size") ||
		strings.Contains(msg, "size not supported") ||
		strings.Contains(msg, "size is not supported") ||
		strings.Contains(msg, "must be one of")
}

var _ ports.ImageGenerator = (*Client)(nil)
