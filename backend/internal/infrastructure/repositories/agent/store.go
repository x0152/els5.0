package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/agent"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) GetOrCreateSession(ctx context.Context, accountID string) (agent.Session, error) {
	sess, err := s.scanSession(s.pool.QueryRow(ctx,
		`SELECT id, account_id, model, context_started_at, created_at, updated_at FROM ai_sessions WHERE account_id = $1`, accountID))
	if err == nil {
		return sess, nil
	}
	if err != pgx.ErrNoRows {
		return agent.Session{}, fmt.Errorf("get session: %w", err)
	}
	now := time.Now().UTC()
	sess = agent.Session{ID: uuid.NewString(), AccountID: accountID, ContextStartedAt: now, CreatedAt: now, UpdatedAt: now}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO ai_sessions (id, account_id, model, context_started_at, created_at, updated_at)
		 VALUES ($1,$2,'',$3,$4,$5) ON CONFLICT (account_id) DO NOTHING`,
		sess.ID, accountID, now, now, now)
	if err != nil {
		return agent.Session{}, fmt.Errorf("create session: %w", err)
	}
	return s.scanSession(s.pool.QueryRow(ctx,
		`SELECT id, account_id, model, context_started_at, created_at, updated_at FROM ai_sessions WHERE account_id = $1`, accountID))
}

func (s *Store) UpdateModel(ctx context.Context, sessionID, model string) error {
	_, err := s.pool.Exec(ctx, `UPDATE ai_sessions SET model=$2, updated_at=now() WHERE id=$1`, sessionID, model)
	if err != nil {
		return fmt.Errorf("update model: %w", err)
	}
	return nil
}

func (s *Store) ResetContext(ctx context.Context, sessionID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE ai_sessions SET context_started_at=$2, updated_at=now() WHERE id=$1`, sessionID, at)
	if err != nil {
		return fmt.Errorf("reset context: %w", err)
	}
	return nil
}

func (s *Store) DeleteMessages(ctx context.Context, sessionID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM ai_messages WHERE session_id=$1`, sessionID)
	if err != nil {
		return fmt.Errorf("delete messages: %w", err)
	}
	return nil
}

func (s *Store) DeleteMessagesFrom(ctx context.Context, sessionID string, from time.Time) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM ai_messages WHERE session_id=$1 AND created_at >= $2`, sessionID, from)
	if err != nil {
		return fmt.Errorf("delete messages from: %w", err)
	}
	return nil
}

func (s *Store) InsertMessage(ctx context.Context, m agent.Message) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	calls, err := json.Marshal(m.ToolCalls)
	if err != nil {
		return fmt.Errorf("marshal tool calls: %w", err)
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO ai_messages (id, session_id, role, content, tool_calls, tool_call_id, tool_name, model, finish_reason, reasoning_content, prompt_tokens, completion_tokens, total_tokens, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		m.ID, m.SessionID, m.Role, m.Content, calls, m.ToolCallID, m.ToolName, m.Model, m.FinishReason, m.ReasoningContent,
		m.PromptTokens, m.CompletionTokens, m.TotalTokens, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
}

func (s *Store) ListMessages(ctx context.Context, sessionID string) ([]agent.Message, error) {
	return s.query(ctx, `WHERE session_id=$1 ORDER BY created_at ASC`, sessionID)
}

func (s *Store) ListMessagesSince(ctx context.Context, sessionID string, since time.Time) ([]agent.Message, error) {
	return s.query(ctx, `WHERE session_id=$1 AND created_at >= $2 ORDER BY created_at ASC`, sessionID, since)
}

func (s *Store) query(ctx context.Context, where string, args ...any) ([]agent.Message, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, session_id, role, content, tool_calls, tool_call_id, tool_name, model, finish_reason, reasoning_content, prompt_tokens, completion_tokens, total_tokens, created_at FROM ai_messages `+where, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()
	out := []agent.Message{}
	for rows.Next() {
		var m agent.Message
		var calls []byte
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &calls, &m.ToolCallID, &m.ToolName, &m.Model, &m.FinishReason, &m.ReasoningContent, &m.PromptTokens, &m.CompletionTokens, &m.TotalTokens, &m.CreatedAt); err != nil {
			return nil, err
		}
		if len(calls) > 0 {
			_ = json.Unmarshal(calls, &m.ToolCalls)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) scanSession(row pgx.Row) (agent.Session, error) {
	var sess agent.Session
	if err := row.Scan(&sess.ID, &sess.AccountID, &sess.Model, &sess.ContextStartedAt, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
		return agent.Session{}, err
	}
	return sess, nil
}
