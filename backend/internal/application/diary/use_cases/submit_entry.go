package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/timex"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type EventSink interface {
	InsertRaw(ctx context.Context, e core.RawEvent) error
}

type SubmitEntryCommand struct {
	Text     string
	Question string
}

type SubmitEntryUseCase struct {
	repo   diary.Repository
	llm    LLMClient
	events EventSink
	clock  timex.Clock
}

func NewSubmitEntryUseCase(repo diary.Repository, llm LLMClient, events EventSink, clock timex.Clock) *SubmitEntryUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &SubmitEntryUseCase{repo: repo, llm: llm, events: events, clock: clock}
}

func (uc *SubmitEntryUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd SubmitEntryCommand) (diary.Entry, error) {
	// 1. Validate the draft and LLM availability.
	accountID := actor.AccountID().String()
	now := uc.clock.Now().In(timex.MSK)
	today := timex.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0)
	entry := diary.Entry{
		ID:        uuid.NewString(),
		AccountID: accountID,
		Date:      today,
		Question:  cmd.Question,
		Text:      cmd.Text,
		CreatedAt: now,
	}
	if err := entry.Validate(); err != nil {
		return diary.Entry{}, err
	}
	if !uc.llm.Available() {
		return diary.Entry{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Reject a second entry for the same day.
	if _, err := uc.repo.GetByDate(ctx, accountID, today); err == nil {
		return diary.Entry{}, fmt.Errorf("entry for today already exists: %w", shared.ErrConflict)
	} else if !errors.Is(err, shared.ErrNotFound) {
		return diary.Entry{}, err
	}

	// 3. Ask the LLM for the friend reply and corrections, with recent entries as context.
	history, err := uc.repo.Latest(ctx, accountID, 3)
	if err != nil {
		return diary.Entry{}, err
	}
	system, user := diary.BuildReplyPrompt(cmd.Question, cmd.Text, actor.Account().NativeLanguage(), history)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return diary.Entry{}, err
	}
	reply, err := diary.ParseReply(raw)
	if err != nil {
		return diary.Entry{}, err
	}
	entry.Reply = reply.Text
	entry.NextQuestion = reply.NextQuestion
	entry.NativeSample = reply.NativeSample
	entry.Corrections = reply.Corrections

	// 4. Persist the entry.
	if err := uc.repo.Insert(ctx, entry); err != nil {
		return diary.Entry{}, err
	}

	// 5. Publish the entry text into the learn core pipeline (best effort).
	if uc.events != nil {
		event := core.RawEvent{
			ID:     uuid.NewString(),
			UserID: accountID,
			Skill:  core.SkillWriting,
			Text:   cmd.Text,
			Source: json.RawMessage(`{"app":"diary"}`),
		}
		core.Normalize(&event, now)
		_ = uc.events.InsertRaw(ctx, event)
	}

	return entry, nil
}
