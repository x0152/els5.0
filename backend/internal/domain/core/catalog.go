package core

import (
	"strings"
	"time"
)

type Word struct {
	ID        string
	Key       string
	Lemma     string
	POS       string
	Type      string
	CEFR      string
	Frequency float64
	Enriched  bool
	IsStop    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GrammarRule struct {
	ID        string
	Key       string
	ParentKey string
	Title     string
	CEFRLevel string
	Enriched  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NormalizeGrammarKey(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	return strings.Join(strings.Fields(k), "_")
}

func WordKey(lemma, pos string) string {
	l := strings.ToLower(strings.TrimSpace(lemma))
	if l == "" {
		return ""
	}
	p := strings.ToUpper(strings.TrimSpace(pos))
	if p == "" {
		return l
	}
	return l + ":" + p
}

func GrammarParentKey(key string) string {
	k := strings.TrimSpace(key)
	if i := strings.LastIndex(k, ":"); i > 0 {
		return k[:i]
	}
	return ""
}
