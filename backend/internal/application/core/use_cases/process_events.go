package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/els/backend/internal/domain/core"
)

type EventStore interface {
	InsertRaw(ctx context.Context, e core.RawEvent) error
	InsertRawBatch(ctx context.Context, events []core.RawEvent) error
	ListRaw(ctx context.Context, userID, status string) ([]core.RawEvent, error)
	ListAllRaw(ctx context.Context, userID string) ([]core.RawEvent, error)
	ClaimPendingRaw(ctx context.Context, limit int) ([]core.RawEvent, error)
	ClaimFailedRaw(ctx context.Context, limit int) ([]core.RawEvent, error)
	RequeueStaleProcessing(ctx context.Context, olderThan time.Duration) error
	ListEvents(ctx context.Context, userID string) ([]core.Event, error)
	ListWords(ctx context.Context, limit int) ([]core.Word, error)
	ListGrammarRules(ctx context.Context, limit int) ([]core.GrammarRule, error)
	Complete(ctx context.Context, rawID string, events []core.Event) error
	Fail(ctx context.Context, rawID, reason string) error
}

const staleProcessingAge = 10 * time.Minute

type LLM interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type ProcessingGate interface {
	Enabled(ctx context.Context) (bool, error)
}

func gateOpen(ctx context.Context, gate ProcessingGate) (bool, error) {
	if gate == nil {
		return true, nil
	}
	return gate.Enabled(ctx)
}

type ProcessEventsUseCase struct {
	store EventStore
	llm   LLM
	gate  ProcessingGate
	log   *slog.Logger
	batch int
}

func NewProcessEventsUseCase(store EventStore, llm LLM, gate ProcessingGate, log *slog.Logger, batch int) *ProcessEventsUseCase {
	if batch <= 0 {
		batch = 5
	}
	return &ProcessEventsUseCase{store: store, llm: llm, gate: gate, log: log, batch: batch}
}

func (uc *ProcessEventsUseCase) Execute(ctx context.Context) error {
	// 1. Event processing is disabled — leave them in pending.
	if on, err := gateOpen(ctx, uc.gate); err != nil || !on {
		return err
	}

	// 2. Re-queue events stuck after an instance crash.
	if err := uc.store.RequeueStaleProcessing(ctx, staleProcessingAge); err != nil {
		return err
	}

	// 3. Claim a batch of unprocessed events exclusively for this instance.
	raws, err := uc.store.ClaimPendingRaw(ctx, uc.batch)
	if err != nil {
		return err
	}

	// 4. Parse and persist.
	processRaws(ctx, uc.store, uc.llm, uc.log, raws)
	return nil
}

func processRaws(ctx context.Context, store EventStore, llm LLM, log *slog.Logger, raws []core.RawEvent) {
	now := time.Now()
	result := make(map[string][]core.Event, len(raws))
	groups := map[string][]core.RawEvent{}
	for _, raw := range raws {
		if raw.IsTargeted() {
			result[raw.ID] = []core.Event{core.TargetedEvent(raw, now)}
			continue
		}
		groups[raw.Skill] = append(groups[raw.Skill], raw)
	}

	for skill, group := range groups {
		if !llm.Available() {
			failGroup(ctx, store, log, group, "llm unavailable")
			continue
		}
		var registry []core.GrammarRule
		if core.WantsGrammarRegistry(skill) {
			var err error
			if registry, err = store.ListGrammarRules(ctx, 500); err != nil {
				failGroup(ctx, store, log, group, err.Error())
				continue
			}
		}

		sys, usr := core.BuildExtractionPrompt(skill, group, registry)
		items, err := extract(ctx, llm, sys, usr)
		if err != nil {
			failGroup(ctx, store, log, group, err.Error())
			continue
		}

		var constructions map[int][]core.Extraction
		if core.WantsConstructions(skill) {
			sys, usr = core.BuildConstructionsPrompt(skill, group, registry)
			if constructions, err = extract(ctx, llm, sys, usr); err != nil {
				failGroup(ctx, store, log, group, err.Error())
				continue
			}
		}

		for i, raw := range group {
			events := core.EventsFromExtractions(raw, items[i], now)
			result[raw.ID] = append(events, core.ConstructionEventsFromExtractions(raw, constructions[i], now)...)
		}
	}

	for id, events := range result {
		if err := store.Complete(ctx, id, events); err != nil {
			log.Error("complete event failed", slog.String("id", id), slog.String("err", err.Error()))
			fail(ctx, store, log, id, err.Error())
		}
	}
}

func extract(ctx context.Context, llm LLM, system, user string) (map[int][]core.Extraction, error) {
	out, err := llm.Chat(ctx, system, user)
	if err != nil {
		return nil, err
	}
	return core.ParseExtractions(out)
}

func failGroup(ctx context.Context, store EventStore, log *slog.Logger, group []core.RawEvent, reason string) {
	for _, raw := range group {
		fail(ctx, store, log, raw.ID, reason)
	}
}

func fail(ctx context.Context, store EventStore, log *slog.Logger, rawID, reason string) {
	if err := store.Fail(ctx, rawID, reason); err != nil {
		log.Error("fail raw event failed", slog.String("id", rawID), slog.String("err", err.Error()))
	}
}
