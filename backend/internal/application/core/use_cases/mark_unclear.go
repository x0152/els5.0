package usecases

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/timex"
)

type UnclearStore interface {
	FindRawByText(ctx context.Context, userID, skill, text string) (core.RawEvent, bool, error)
	SetRawOutcome(ctx context.Context, rawID, outcome string) error
	InsertRaw(ctx context.Context, e core.RawEvent) error
}

type MarkUnclearUseCase struct {
	store UnclearStore
	clock timex.Clock
}

func NewMarkUnclearUseCase(store UnclearStore, clock timex.Clock) *MarkUnclearUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &MarkUnclearUseCase{store: store, clock: clock}
}

type MarkUnclearCommand struct {
	Event core.RawEvent
}

type MarkUnclearResult struct {
	Updated bool
}

func (uc *MarkUnclearUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd MarkUnclearCommand) (MarkUnclearResult, error) {
	if actor == nil {
		return MarkUnclearResult{}, shared.ErrUnauthorized
	}

	now := uc.clock.Now()
	e := cmd.Event
	core.Normalize(&e, now)
	e.Text = strings.TrimSpace(e.Text)
	e.Outcome = "fail"
	if err := e.Validate(); err != nil {
		return MarkUnclearResult{}, err
	}

	userID := actor.AccountID().String()
	if existing, found, err := uc.store.FindRawByText(ctx, userID, e.Skill, e.Text); err != nil {
		return MarkUnclearResult{}, err
	} else if found {
		if err := uc.store.SetRawOutcome(ctx, existing.ID, "fail"); err != nil {
			return MarkUnclearResult{}, err
		}
		return MarkUnclearResult{Updated: true}, nil
	}

	e.ID = uuid.NewString()
	e.UserID = userID
	if err := uc.store.InsertRaw(ctx, e); err != nil {
		return MarkUnclearResult{}, err
	}
	return MarkUnclearResult{Updated: false}, nil
}
