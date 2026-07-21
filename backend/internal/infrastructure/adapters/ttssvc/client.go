package ttssvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 120 * time.Second}}
}

func (c *Client) Synthesize(ctx context.Context, text, voice string, speed float64) ([]byte, error) {
	body, err := json.Marshal(map[string]any{"text": text, "voice": voice, "speed": speed})
	if err != nil {
		return nil, fmt.Errorf("marshal tts request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/tts", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tts service request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tts service returned %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
