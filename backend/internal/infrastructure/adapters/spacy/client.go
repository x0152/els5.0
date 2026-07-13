package spacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/els/backend/internal/domain/lexicon"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{}}
}

type spanDTO struct {
	Position int    `json:"position"`
	SpanType string `json:"span_type"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Text     string `json:"text"`
}

type unitDTO struct {
	UnitType    string         `json:"unit_type"`
	BaseForm    string         `json:"base_form"`
	POS         *string        `json:"pos"`
	SentenceIdx int            `json:"sentence_idx"`
	Metadata    map[string]any `json:"metadata"`
	Language    string         `json:"language"`
	Spans       []spanDTO      `json:"spans"`
}

type baseFormDTO struct {
	Text     string  `json:"text"`
	POS      *string `json:"pos"`
	IsStop   bool    `json:"is_stop"`
	Language string  `json:"language"`
}

type analyzeDTO struct {
	SentenceCount int                    `json:"sentence_count"`
	Language      string                 `json:"language"`
	Units         []unitDTO              `json:"units"`
	BaseForms     map[string]baseFormDTO `json:"base_forms"`
}

func (c *Client) Analyze(ctx context.Context, html string) (lexicon.Analysis, error) {
	body, err := json.Marshal(map[string]string{"html_content": html})
	if err != nil {
		return lexicon.Analysis{}, fmt.Errorf("marshal analyze request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analyze", bytes.NewReader(body))
	if err != nil {
		return lexicon.Analysis{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return lexicon.Analysis{}, fmt.Errorf("spacy request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return lexicon.Analysis{}, fmt.Errorf("spacy returned %d", resp.StatusCode)
	}
	var dto analyzeDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return lexicon.Analysis{}, fmt.Errorf("decode analyze response: %w", err)
	}
	return dto.toDomain(), nil
}

func (d analyzeDTO) toDomain() lexicon.Analysis {
	a := lexicon.Analysis{
		SentenceCount: d.SentenceCount,
		Language:      d.Language,
		Units:         make([]lexicon.Unit, 0, len(d.Units)),
		BaseForms:     make(map[string]lexicon.BaseForm, len(d.BaseForms)),
	}
	for key, bf := range d.BaseForms {
		a.BaseForms[key] = lexicon.BaseForm{
			Text:     bf.Text,
			POS:      deref(bf.POS),
			IsStop:   bf.IsStop,
			Language: bf.Language,
		}
	}
	for _, u := range d.Units {
		spans := make([]lexicon.Span, 0, len(u.Spans))
		for _, s := range u.Spans {
			spans = append(spans, lexicon.Span{
				Position: s.Position,
				SpanType: s.SpanType,
				Start:    s.Start,
				End:      s.End,
				Text:     s.Text,
			})
		}
		a.Units = append(a.Units, lexicon.Unit{
			UnitType:    lexicon.UnitType(u.UnitType),
			BaseForm:    u.BaseForm,
			POS:         deref(u.POS),
			SentenceIdx: u.SentenceIdx,
			Metadata:    u.Metadata,
			Language:    u.Language,
			Spans:       spans,
		})
	}
	return a
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
