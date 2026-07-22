package usecases

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/els/backend/internal/domain/agent"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

type Service struct {
	repo    agent.Repository
	loop    *agent.Loop
	llm     agent.LLM
	running sync.Map
}

func NewService(repo agent.Repository, loop *agent.Loop, llm agent.LLM) *Service {
	return &Service{repo: repo, loop: loop, llm: llm}
}

type HistoryResult struct {
	Model        string
	DefaultModel string
	Generating   bool
	Messages     []agent.Message
}

func (s *Service) History(ctx context.Context, actor *iam.Actor) (HistoryResult, error) {
	sess, err := s.session(ctx, actor)
	if err != nil {
		return HistoryResult{}, err
	}
	msgs, err := s.repo.ListMessages(ctx, sess.ID)
	if err != nil {
		return HistoryResult{}, err
	}
	_, generating := s.running.Load(sess.ID)
	return HistoryResult{Model: s.modelOf(sess), DefaultModel: s.llm.DefaultModel(), Generating: generating, Messages: msgs}, nil
}

type ModelsResult struct {
	Models   []agent.LLMModel
	Selected string
	Default  string
}

func (s *Service) Models(ctx context.Context, actor *iam.Actor) (ModelsResult, error) {
	sess, err := s.session(ctx, actor)
	if err != nil {
		return ModelsResult{}, err
	}
	models, err := s.llm.ListModels(ctx)
	if err != nil {
		return ModelsResult{}, err
	}
	return ModelsResult{Models: models, Selected: s.modelOf(sess), Default: s.llm.DefaultModel()}, nil
}

func (s *Service) SetModel(ctx context.Context, actor *iam.Actor, model string) error {
	sess, err := s.session(ctx, actor)
	if err != nil {
		return err
	}
	return s.repo.UpdateModel(ctx, sess.ID, model)
}

func (s *Service) Reset(ctx context.Context, actor *iam.Actor) error {
	sess, err := s.session(ctx, actor)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	// The separator stays in the feed as a delimiter, and the new LLM context starts from now.
	if err := s.repo.InsertMessage(ctx, agent.Message{SessionID: sess.ID, Role: agent.RoleSeparator, CreatedAt: now}); err != nil {
		return err
	}
	return s.repo.ResetContext(ctx, sess.ID, now)
}

// FillGap stores the user's answer inside the assistant message that contains
// the exercise, so later agent runs see the fills in their own history.
func (s *Service) FillGap(ctx context.Context, actor *iam.Actor, messageID string, ordinal int, answer string) error {
	// 1. Resolve the actor's session.
	sess, err := s.session(ctx, actor)
	if err != nil {
		return err
	}
	// 2. Load the message and make sure it is this user's assistant message.
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.SessionID != sess.ID || msg.Role != agent.RoleAssistant {
		return shared.ErrNotFound
	}
	// 3. Write the fill into the gap.
	content, ok := agent.FillGap(msg.Content, ordinal, answer)
	if !ok {
		return fmt.Errorf("%w: gap %d not found", shared.ErrValidation, ordinal)
	}
	return s.repo.UpdateMessageContent(ctx, msg.ID, content)
}

func (s *Service) Clear(ctx context.Context, actor *iam.Actor) error {
	sess, err := s.session(ctx, actor)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteMessages(ctx, sess.ID); err != nil {
		return err
	}
	return s.repo.ResetContext(ctx, sess.ID, time.Now().UTC())
}

// Stream runs the request through the agent loop, streams events via emit, and
// at the same time persists the dialogue turn in the DB for subsequent requests.
// Generation is not tied to the HTTP request lifetime: if the connection drops
// (page refresh) the loop still finishes and saves the reply to history.
func (s *Service) Stream(ctx context.Context, actor *iam.Actor, userMessage string, view *agent.View, emit func(agent.Event)) error {
	sess, err := s.session(ctx, actor)
	if err != nil {
		return err
	}
	unlock, err := s.tryLock(sess.ID)
	if err != nil {
		return err
	}
	defer unlock()
	return s.stream(ctx, actor, sess, userMessage, view, emit)
}

func (s *Service) stream(ctx context.Context, actor *iam.Actor, sess agent.Session, userMessage string, view *agent.View, emit func(agent.Event)) error {
	ctx = context.WithoutCancel(ctx)
	history, err := s.repo.ListMessagesSince(ctx, sess.ID, sess.ContextStartedAt)
	if err != nil {
		return err
	}
	if err := s.repo.InsertMessage(ctx, agent.Message{SessionID: sess.ID, Role: agent.RoleUser, Content: userMessage, CreatedAt: time.Now().UTC()}); err != nil {
		return err
	}

	ch, err := s.loop.Run(agent.WithActor(ctx, actor), agent.RunContext{
		Actor:       actor,
		History:     history,
		UserMessage: userMessage,
		Model:       s.modelOf(sess),
		View:        view,
	})
	if err != nil {
		return err
	}

	// The channel is drained to the end even after a save error,
	// so the agent loop can finish cleanly; the error is returned afterward.
	var saveErr error
	save := func(m agent.Message) {
		if err := s.repo.InsertMessage(ctx, m); err != nil && saveErr == nil {
			saveErr = fmt.Errorf("save chat message: %w", err)
		}
	}
	for ev := range ch {
		emit(ev)
		switch ev.Type {
		case agent.EventAssistantTurn:
			save(agent.Message{
				SessionID:        sess.ID,
				Role:             agent.RoleAssistant,
				Content:          ev.Text,
				ToolCalls:        ev.ToolCalls,
				Model:            ev.Model,
				FinishReason:     string(ev.FinishReason),
				PromptTokens:     ev.Usage.PromptTokens,
				CompletionTokens: ev.Usage.CompletionTokens,
				TotalTokens:      ev.Usage.TotalTokens,
				CreatedAt:        time.Now().UTC(),
			})
		case agent.EventToolEnd:
			if ev.Step != nil {
				save(agent.Message{
					SessionID:  sess.ID,
					Role:       agent.RoleTool,
					Content:    ev.ToolResult,
					ToolCallID: ev.Step.ID,
					ToolName:   ev.Step.Tool,
					CreatedAt:  time.Now().UTC(),
				})
			}
		}
	}
	return saveErr
}

// Regenerate removes the last exchange (user request and reply) and re-runs
// the same user request through the agent loop.
func (s *Service) Regenerate(ctx context.Context, actor *iam.Actor, view *agent.View, emit func(agent.Event)) error {
	sess, err := s.session(ctx, actor)
	if err != nil {
		return err
	}
	unlock, err := s.tryLock(sess.ID)
	if err != nil {
		return err
	}
	defer unlock()
	msgs, err := s.repo.ListMessagesSince(ctx, sess.ID, sess.ContextStartedAt)
	if err != nil {
		return err
	}
	var last *agent.Message
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == agent.RoleUser {
			last = &msgs[i]
			break
		}
	}
	if last == nil {
		return nil
	}
	if err := s.repo.DeleteMessagesFrom(ctx, sess.ID, last.CreatedAt); err != nil {
		return err
	}
	return s.stream(ctx, actor, sess, last.Content, view, emit)
}

func (s *Service) tryLock(sessionID string) (func(), error) {
	if _, loaded := s.running.LoadOrStore(sessionID, true); loaded {
		return nil, fmt.Errorf("%w: generation already running", shared.ErrConflict)
	}
	return func() { s.running.Delete(sessionID) }, nil
}

func (s *Service) session(ctx context.Context, actor *iam.Actor) (agent.Session, error) {
	if actor == nil {
		return agent.Session{}, shared.ErrUnauthorized
	}
	return s.repo.GetOrCreateSession(ctx, actor.AccountID().String())
}

func (s *Service) modelOf(sess agent.Session) string {
	if sess.Model != "" {
		return sess.Model
	}
	return s.llm.DefaultModel()
}
