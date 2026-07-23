package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/studio"
	"github.com/els/backend/internal/utils/timex"
)

type EventSink interface {
	InsertRaw(ctx context.Context, e core.RawEvent) error
}

func emitItemEvent(ctx context.Context, sink EventSink, now time.Time, userID string, item studio.Item, skill, outcome, usage string) {
	if sink == nil {
		return
	}
	e := core.RawEvent{
		ID:      uuid.NewString(),
		UserID:  userID,
		Skill:   skill,
		Target:  item.Text,
		Outcome: outcome,
		Context: usage,
		Source:  json.RawMessage(`{"app":"studio"}`),
	}
	core.Normalize(&e, now)
	_ = sink.InsertRaw(ctx, e)
}

type CheckReplyCommand struct {
	ItemID string
	Reply  string
}

type CheckReplyUseCase struct {
	repo   studio.Repository
	llm    LLMClient
	events EventSink
	clock  timex.Clock
}

func NewCheckReplyUseCase(repo studio.Repository, llm LLMClient, events EventSink, clock timex.Clock) *CheckReplyUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &CheckReplyUseCase{repo: repo, llm: llm, events: events, clock: clock}
}

func (uc *CheckReplyUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd CheckReplyCommand) (studio.ReplyCheck, error) {
	// 1. Load the item and validate the state.
	item, err := uc.repo.Get(ctx, actor.AccountID().String(), cmd.ItemID)
	if err != nil {
		return studio.ReplyCheck{}, err
	}
	if item.Task == "" {
		return studio.ReplyCheck{}, fmt.Errorf("task is not generated yet: %w", shared.ErrValidation)
	}
	if !uc.llm.Available() {
		return studio.ReplyCheck{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Ask the LLM to judge the reply.
	system, user := studio.BuildCheckPrompt(item.Text, item.Task, cmd.Reply)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return studio.ReplyCheck{}, err
	}
	check, err := studio.ParseCheck(raw)
	if err != nil {
		return studio.ReplyCheck{}, err
	}

	// 3. A passing reply marks the written skill and may schedule the first review.
	if check.OK && !item.Written {
		item.Written = true
		item.ScheduleReviewIfDone(uc.clock.Now())
		if err := uc.repo.Update(ctx, item); err != nil {
			return studio.ReplyCheck{}, err
		}
	}

	// 4. Publish the attempt into the learn core pipeline (best effort).
	outcome := "fail"
	if check.OK {
		outcome = "ok"
	}
	emitItemEvent(ctx, uc.events, uc.clock.Now(), actor.AccountID().String(), item, core.SkillWriting, outcome, cmd.Reply)

	return check, nil
}
