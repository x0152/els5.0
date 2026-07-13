package usecases

import (
	"context"
	"log/slog"
)

type RetryFailedEventsUseCase struct {
	store EventStore
	llm   LLM
	gate  ProcessingGate
	log   *slog.Logger
	batch int
}

func NewRetryFailedEventsUseCase(store EventStore, llm LLM, gate ProcessingGate, log *slog.Logger, batch int) *RetryFailedEventsUseCase {
	if batch <= 0 {
		batch = 5
	}
	return &RetryFailedEventsUseCase{store: store, llm: llm, gate: gate, log: log, batch: batch}
}

func (uc *RetryFailedEventsUseCase) Execute(ctx context.Context) error {
	// 1. Event processing is disabled — do not retry anything.
	if on, err := gateOpen(ctx, uc.gate); err != nil || !on {
		return err
	}

	// 2. Fetch events that failed on previous ticks (e.g. due to bad JSON from the LLM).
	raws, err := uc.store.ClaimFailedRaw(ctx, uc.batch)
	if err != nil {
		return err
	}

	// 3. Re-run parsing and persist; ones that fail again stay failed until the next tick.
	processRaws(ctx, uc.store, uc.llm, uc.log, raws)
	return nil
}
