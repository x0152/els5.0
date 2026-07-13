package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/timex"
)

type IngestEventsUseCase struct {
	store EventStore
	clock timex.Clock
}

func NewIngestEventsUseCase(store EventStore, clock timex.Clock) *IngestEventsUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &IngestEventsUseCase{store: store, clock: clock}
}

type IngestEventsCommand struct {
	Events []core.RawEvent
}

type IngestEventsResult struct {
	Accepted int
}

func (uc *IngestEventsUseCase) Execute(
	ctx context.Context,
	actor *iam.Actor,
	cmd IngestEventsCommand,
) (IngestEventsResult, error) {
	// 1. Only an authenticated actor may send events.
	if actor == nil {
		return IngestEventsResult{}, shared.ErrUnauthorized
	}

	// 2. Normalize, validate, and set the owner on each event.
	now := uc.clock.Now()
	userID := actor.AccountID().String()
	for i := range cmd.Events {
		core.Normalize(&cmd.Events[i], now)
		if err := cmd.Events[i].Validate(); err != nil {
			return IngestEventsResult{}, err
		}
		cmd.Events[i].ID = uuid.NewString()
		cmd.Events[i].UserID = userID
	}

	// 3. Persist the whole batch atomically — the background Processor will pick it up.
	if err := uc.store.InsertRawBatch(ctx, cmd.Events); err != nil {
		return IngestEventsResult{}, err
	}

	// 4. Return the number of accepted events.
	return IngestEventsResult{Accepted: len(cmd.Events)}, nil
}
