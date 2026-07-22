package usecases

import (
	"context"
	"sync"
	"time"

	"github.com/els/backend/internal/domain/diary"
)

// ReplyWorker generates the friend reply in the background so the user
// can close the app right after submitting. A pending entry is re-kicked
// on the next visit if the previous attempt failed or the server restarted.
type ReplyWorker struct {
	repo    diary.Repository
	llm     LLMClient
	running sync.Map
}

func NewReplyWorker(repo diary.Repository, llm LLMClient) *ReplyWorker {
	return &ReplyWorker{repo: repo, llm: llm}
}

func (w *ReplyWorker) Kick(entry diary.Entry, nativeLanguage string) {
	if _, loaded := w.running.LoadOrStore(entry.ID, true); loaded {
		return
	}
	go func() {
		defer w.running.Delete(entry.ID)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		recent, err := w.repo.Latest(ctx, entry.AccountID, 4)
		if err != nil {
			return
		}
		history := make([]diary.Entry, 0, 3)
		for _, e := range recent {
			if e.ID != entry.ID && len(history) < 3 {
				history = append(history, e)
			}
		}

		system, user := diary.BuildReplyPrompt(entry.Question, entry.Text, nativeLanguage, history)
		raw, err := w.llm.Chat(ctx, system, user)
		if err != nil {
			return
		}
		reply, err := diary.ParseReply(raw)
		if err != nil {
			return
		}
		entry.Reply = reply.Text
		entry.NextQuestion = reply.NextQuestion
		entry.NativeSample = reply.NativeSample
		entry.Corrections = reply.Corrections
		entry.Status = diary.StatusDone
		_ = w.repo.UpdateReply(ctx, entry)
	}()
}
