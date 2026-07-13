package core

import (
	"encoding/json"
	"fmt"
)

const wordEnrichSystem = `You enrich a BATCH of vocabulary entries (lemma + part of speech).
Input is a JSON array of items, each with an integer "index", "lemma", "pos".
Return ONLY one JSON object: {"results": [{"index": <int>, "cefr": "A1|A2|B1|B2|C1|C2", "frequency": <number 0..1>}]} with exactly one entry per input item, matched by its index.
- cefr: the CEFR level at which a learner typically meets this word.
- frequency: how common the word is in everyday English, from 0 (rare) to 1 (very common).`

const grammarEnrichSystem = `You enrich a BATCH of grammar rules identified by a key like "tense:present_perfect:active".
Input is a JSON array of items, each with an integer "index", "key", "parent_key".
Return ONLY one JSON object: {"results": [{"index": <int>, "title": string, "cefr_level": "A1|A2|B1|B2|C1|C2"}]} with exactly one entry per input item, matched by its index.
- title: a short human-readable name of the rule, in English only.
- cefr_level: the CEFR level at which this rule is typically taught.`

type WordEnrichment struct {
	CEFR      string  `json:"cefr"`
	Frequency float64 `json:"frequency"`
}

type GrammarEnrichment struct {
	Title     string `json:"title"`
	CEFRLevel string `json:"cefr_level"`
}

func BuildWordEnrichmentPrompt(words []Word) (system, user string) {
	type in struct {
		Index int    `json:"index"`
		Lemma string `json:"lemma"`
		POS   string `json:"pos"`
	}
	items := make([]in, len(words))
	for i, w := range words {
		items[i] = in{Index: i, Lemma: w.Lemma, POS: w.POS}
	}
	b, _ := json.Marshal(items)
	return wordEnrichSystem, string(b)
}

func ParseWordEnrichments(out string) (map[int]WordEnrichment, error) {
	var parsed struct {
		Results []struct {
			Index int `json:"index"`
			WordEnrichment
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return nil, fmt.Errorf("parse word enrichment: %w", err)
	}
	m := make(map[int]WordEnrichment, len(parsed.Results))
	for _, r := range parsed.Results {
		m[r.Index] = r.WordEnrichment
	}
	return m, nil
}

func BuildGrammarEnrichmentPrompt(rules []GrammarRule) (system, user string) {
	type in struct {
		Index     int    `json:"index"`
		Key       string `json:"key"`
		ParentKey string `json:"parent_key"`
	}
	items := make([]in, len(rules))
	for i, r := range rules {
		items[i] = in{Index: i, Key: r.Key, ParentKey: r.ParentKey}
	}
	b, _ := json.Marshal(items)
	return grammarEnrichSystem, string(b)
}

func ParseGrammarEnrichments(out string) (map[int]GrammarEnrichment, error) {
	var parsed struct {
		Results []struct {
			Index int `json:"index"`
			GrammarEnrichment
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return nil, fmt.Errorf("parse grammar enrichment: %w", err)
	}
	m := make(map[int]GrammarEnrichment, len(parsed.Results))
	for _, r := range parsed.Results {
		m[r.Index] = r.GrammarEnrichment
	}
	return m, nil
}
