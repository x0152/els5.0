package speechsvc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/els/backend/internal/domain/speech"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 120 * time.Second}}
}

type phonemeDTO struct {
	Expected string  `json:"expected"`
	Heard    string  `json:"heard"`
	Score    float64 `json:"score"`
	Verdict  string  `json:"verdict"`
}

type wordDTO struct {
	Word     string       `json:"word"`
	IPA      string       `json:"ipa"`
	Score    int          `json:"score"`
	Phonemes []phonemeDTO `json:"phonemes"`
	Extra    []string     `json:"extra"`
}

type assessDTO struct {
	Error   string    `json:"error"`
	Overall int       `json:"overall"`
	Heard   string    `json:"heard"`
	Words   []wordDTO `json:"words"`
}

func (c *Client) Assess(ctx context.Context, audio []byte, text string, strictness float64) (speech.Assessment, error) {
	body, err := json.Marshal(map[string]any{
		"audio_base64": base64.StdEncoding.EncodeToString(audio),
		"text":         text,
		"strictness":   strictness,
	})
	if err != nil {
		return speech.Assessment{}, fmt.Errorf("marshal assess request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/assess", bytes.NewReader(body))
	if err != nil {
		return speech.Assessment{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return speech.Assessment{}, fmt.Errorf("speech service request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return speech.Assessment{}, fmt.Errorf("speech service returned %d", resp.StatusCode)
	}
	var dto assessDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return speech.Assessment{}, fmt.Errorf("decode assess response: %w", err)
	}
	if dto.Error != "" {
		return speech.Assessment{}, fmt.Errorf("speech service: %s", dto.Error)
	}
	return dto.toDomain(), nil
}

func (d assessDTO) toDomain() speech.Assessment {
	a := speech.Assessment{
		Overall: d.Overall,
		Heard:   d.Heard,
		Words:   make([]speech.Word, 0, len(d.Words)),
	}
	for _, w := range d.Words {
		phonemes := make([]speech.Phoneme, 0, len(w.Phonemes))
		for _, p := range w.Phonemes {
			phonemes = append(phonemes, speech.Phoneme{
				Expected: p.Expected,
				Heard:    p.Heard,
				Score:    p.Score,
				Verdict:  speech.Verdict(p.Verdict),
			})
		}
		extra := w.Extra
		if extra == nil {
			extra = []string{}
		}
		a.Words = append(a.Words, speech.Word{
			Word:     w.Word,
			IPA:      w.IPA,
			Score:    w.Score,
			Phonemes: phonemes,
			Extra:    extra,
		})
	}
	return a
}
