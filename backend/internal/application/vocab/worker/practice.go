package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type LLM interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type Practice struct {
	sessions vocab.PracticeSessionRepository
	llm      LLM
	logger   *slog.Logger
}

func NewPractice(sessions vocab.PracticeSessionRepository, llm LLM, logger *slog.Logger) *Practice {
	if logger == nil {
		logger = slog.Default()
	}
	return &Practice{sessions: sessions, llm: llm, logger: logger}
}

func (w *Practice) Enqueue(accountID, sessionID string, units []vocab.Unit) {
	go w.run(accountID, sessionID, units)
}

func (w *Practice) run(accountID, sessionID string, units []vocab.Unit) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error("vocab: practice generation panic", slog.String("session", sessionID), slog.Any("panic", r))
			w.fail(sessionID, accountID, fmt.Errorf("internal error during generation"))
		}
	}()

	if !w.llm.Available() {
		w.fail(sessionID, accountID, fmt.Errorf("generation is not available"))
		return
	}

	matchSys, matchUser := vocab.BuildMatchPrompt(units)
	match, err := w.stage(ctx, matchSys, matchUser)
	if err != nil {
		w.fail(sessionID, accountID, err)
		return
	}
	if !w.append(ctx, accountID, sessionID, match, vocab.PracticeStatusGenerating) {
		return
	}

	gapSys, gapUser := vocab.BuildGapPrompt(units)
	gap, err := w.stage(ctx, gapSys, gapUser)
	if err != nil {
		w.fail(sessionID, accountID, err)
		return
	}
	if !w.append(ctx, accountID, sessionID, gap, vocab.PracticeStatusGenerating) {
		return
	}

	writes := vocab.BuildWriteExercises(units, 3)
	w.append(ctx, accountID, sessionID, writes, vocab.PracticeStatusReady)
}

// append saves a batch of exercises; the session may have been replaced by a new one (ErrNotFound) —
// then generation simply stops; otherwise the session is marked with an error.
func (w *Practice) append(ctx context.Context, accountID, sessionID, chunk, status string) bool {
	err := w.sessions.AppendExercises(ctx, accountID, sessionID, chunk, status, "")
	if err == nil {
		return true
	}
	if errors.Is(err, shared.ErrNotFound) {
		w.logger.Info("vocab: practice session gone, generation stopped", slog.String("session", sessionID))
		return false
	}
	w.logger.Error("vocab: append practice failed", slog.String("session", sessionID), slog.String("err", err.Error()))
	w.fail(sessionID, accountID, fmt.Errorf("failed to save exercises"))
	return false
}

func (w *Practice) stage(ctx context.Context, system, user string) (string, error) {
	raw, err := w.llm.Chat(ctx, system, user)
	if err != nil {
		return "", err
	}
	return vocab.ParseGeneratedPractice(raw)
}

func (w *Practice) fail(sessionID, accountID string, cause error) {
	w.logger.Warn("vocab: practice generation failed", slog.String("session", sessionID), slog.String("err", cause.Error()))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := w.sessions.AppendExercises(ctx, accountID, sessionID, "", vocab.PracticeStatusError, cause.Error())
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		w.logger.Error("vocab: mark practice failed", slog.String("session", sessionID), slog.String("err", err.Error()))
	}
}
