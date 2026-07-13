package usecases

import (
	"context"
	"log/slog"

	"github.com/els/backend/internal/domain/core"
)

type CatalogStore interface {
	ListUnenrichedWords(ctx context.Context, limit int) ([]core.Word, error)
	ListUnenrichedGrammarRules(ctx context.Context, limit int) ([]core.GrammarRule, error)
	EnrichWord(ctx context.Context, id, cefr string, frequency float64) error
	EnrichGrammarRule(ctx context.Context, id, title, cefrLevel string) error
}

type EnrichCatalogUseCase struct {
	store CatalogStore
	llm   LLM
	gate  ProcessingGate
	log   *slog.Logger
	batch int
}

func NewEnrichCatalogUseCase(store CatalogStore, llm LLM, gate ProcessingGate, log *slog.Logger, batch int) *EnrichCatalogUseCase {
	if batch <= 0 {
		batch = 5
	}
	return &EnrichCatalogUseCase{store: store, llm: llm, gate: gate, log: log, batch: batch}
}

func (uc *EnrichCatalogUseCase) Execute(ctx context.Context) error {
	// 1. Processing is disabled or there is no LLM — skip enrichment.
	if on, err := gateOpen(ctx, uc.gate); err != nil || !on {
		return err
	}
	if !uc.llm.Available() {
		return nil
	}

	// 2. Enrich a batch of new words and grammar rules.
	enrichWords(ctx, uc.store, uc.llm, uc.log, uc.batch)
	enrichGrammarRules(ctx, uc.store, uc.llm, uc.log, uc.batch)
	return nil
}

func enrichWords(ctx context.Context, store CatalogStore, llm LLM, log *slog.Logger, batch int) {
	words, err := store.ListUnenrichedWords(ctx, batch)
	if err != nil || len(words) == 0 {
		return
	}
	system, user := core.BuildWordEnrichmentPrompt(words)
	out, err := llm.Chat(ctx, system, user)
	if err != nil {
		log.Error("enrich words: llm chat failed", slog.String("err", err.Error()))
		return
	}
	items, err := core.ParseWordEnrichments(out)
	if err != nil {
		log.Error("enrich words: parse failed", slog.String("err", err.Error()))
		return
	}
	for i, w := range words {
		e, ok := items[i]
		if !ok {
			continue
		}
		if err := store.EnrichWord(ctx, w.ID, e.CEFR, e.Frequency); err != nil {
			log.Error("enrich word: update failed", slog.String("id", w.ID), slog.String("err", err.Error()))
		}
	}
}

func enrichGrammarRules(ctx context.Context, store CatalogStore, llm LLM, log *slog.Logger, batch int) {
	rules, err := store.ListUnenrichedGrammarRules(ctx, batch)
	if err != nil || len(rules) == 0 {
		return
	}
	system, user := core.BuildGrammarEnrichmentPrompt(rules)
	out, err := llm.Chat(ctx, system, user)
	if err != nil {
		log.Error("enrich grammar: llm chat failed", slog.String("err", err.Error()))
		return
	}
	items, err := core.ParseGrammarEnrichments(out)
	if err != nil {
		log.Error("enrich grammar: parse failed", slog.String("err", err.Error()))
		return
	}
	for i, r := range rules {
		e, ok := items[i]
		if !ok {
			continue
		}
		if err := store.EnrichGrammarRule(ctx, r.ID, e.Title, e.CEFRLevel); err != nil {
			log.Error("enrich grammar: update failed", slog.String("id", r.ID), slog.String("err", err.Error()))
		}
	}
}
