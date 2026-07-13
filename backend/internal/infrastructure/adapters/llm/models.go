package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared/ports"
)

type ModelLister struct {
	http *http.Client
}

func NewModelLister() *ModelLister {
	return &ModelLister{http: &http.Client{Timeout: 30 * time.Second}}
}

func (l *ModelLister) ListModels(ctx context.Context, cfg ports.AIProviderConfig) ([]string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("llm: base url not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := l.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm models status %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(out.Data))
	for _, m := range out.Data {
		if id := strings.TrimSpace(m.ID); id != "" {
			models = append(models, id)
		}
	}
	sort.Strings(models)
	return models, nil
}
