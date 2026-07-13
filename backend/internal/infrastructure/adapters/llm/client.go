package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared/ports"
)

type Client struct {
	cfg      ports.AIProviderConfig
	resolver ports.AIProviderResolver
	http     *http.Client
}

func New(baseURL, apiKey, model string, timeout time.Duration) *Client {
	return &Client{
		cfg: ports.AIProviderConfig{
			BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
			APIKey:  strings.TrimSpace(apiKey),
			Model:   strings.TrimSpace(model),
		},
		http: &http.Client{Timeout: timeout},
	}
}

func NewWithResolver(baseURL, apiKey, model string, timeout time.Duration, resolver ports.AIProviderResolver) *Client {
	c := New(baseURL, apiKey, model, timeout)
	c.resolver = resolver
	return c
}

func (c *Client) resolve(ctx context.Context) ports.AIProviderConfig {
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

func (c *Client) Available() bool {
	cfg := c.resolve(context.Background())
	return cfg.APIKey != "" && cfg.Model != ""
}

const maxAttempts = 3

func (c *Client) Chat(ctx context.Context, system, user string) (string, error) {
	cfg := c.resolve(ctx)
	body, _ := json.Marshal(map[string]any{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	})

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
		content, retryable, err := c.chatOnce(ctx, cfg, body)
		if err == nil {
			return content, nil
		}
		lastErr = err
		if !retryable {
			return "", err
		}
	}
	return "", lastErr
}

func (c *Client) chatOnce(ctx context.Context, cfg ports.AIProviderConfig, body []byte) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", ctx.Err() == nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode >= 500, fmt.Errorf("llm status %d: %s", resp.StatusCode, string(data))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", false, err
	}
	if len(out.Choices) == 0 {
		return "", false, fmt.Errorf("llm empty response")
	}
	return clean(out.Choices[0].Message.Content), false, nil
}

func (c *Client) ChatStream(ctx context.Context, system, user string, onDelta func(string)) error {
	cfg := c.resolve(ctx)
	body, _ := json.Marshal(map[string]any{
		"model":  cfg.Model,
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm status %d: %s", resp.StatusCode, string(data))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(payload), &chunk) != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			onDelta(chunk.Choices[0].Delta.Content)
		}
	}
	return scanner.Err()
}

func clean(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
