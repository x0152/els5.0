package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type ListEventsUseCase struct {
	store EventStore
}

func NewListEventsUseCase(store EventStore) *ListEventsUseCase {
	return &ListEventsUseCase{store: store}
}

type ListEventsQuery struct {
	Status core.Status
}

type ListEventsResult struct {
	Status    core.Status
	Raws      []core.RawEvent
	Processed []core.Event
}

func (uc *ListEventsUseCase) Execute(
	ctx context.Context,
	actor *iam.Actor,
	q ListEventsQuery,
) (ListEventsResult, error) {
	// 1. Only an authenticated actor can see their events.
	if actor == nil {
		return ListEventsResult{}, shared.ErrUnauthorized
	}
	userID := actor.AccountID().String()

	// 2. Processed events live in a separate table.
	if q.Status == core.StatusProcessed {
		events, err := uc.store.ListEvents(ctx, userID)
		if err != nil {
			return ListEventsResult{}, err
		}
		return ListEventsResult{Status: core.StatusProcessed, Processed: events}, nil
	}

	// 3. One query for all events: processed + raw (pending/failed).
	if q.Status == core.StatusAll {
		events, err := uc.store.ListEvents(ctx, userID)
		if err != nil {
			return ListEventsResult{}, err
		}
		pending, err := uc.store.ListRaw(ctx, userID, string(core.StatusPending))
		if err != nil {
			return ListEventsResult{}, err
		}
		failed, err := uc.store.ListRaw(ctx, userID, string(core.StatusFailed))
		if err != nil {
			return ListEventsResult{}, err
		}
		return ListEventsResult{Status: core.StatusAll, Processed: events, Raws: append(pending, failed...)}, nil
	}

	// 4. Raw requests in full, as they arrived.
	if q.Status == core.StatusRaw {
		raws, err := uc.store.ListAllRaw(ctx, userID)
		if err != nil {
			return ListEventsResult{}, err
		}
		return ListEventsResult{Status: core.StatusRaw, Raws: raws}, nil
	}

	// 5. Other statuses (pending/failed) live among the raw events.
	raws, err := uc.store.ListRaw(ctx, userID, string(q.Status))
	if err != nil {
		return ListEventsResult{}, err
	}
	return ListEventsResult{Status: q.Status, Raws: raws}, nil
}
