package vocab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

const practiceTTL = 7 * 24 * time.Hour

type PracticeSessionStore struct {
	client *redis.Client
}

func NewPracticeSessionStore(client *redis.Client) *PracticeSessionStore {
	return &PracticeSessionStore{client: client}
}

type practiceContent struct {
	ID        string       `json:"id"`
	Words     []vocab.Unit `json:"words"`
	Exercises string       `json:"exercises"`
	Status    string       `json:"status"`
	Error     string       `json:"error,omitempty"`
}

type practiceProgress struct {
	SessionID string                          `json:"session_id"`
	Answers   map[string]vocab.PracticeAnswer `json:"answers"`
	Completed bool                            `json:"completed"`
}

func (s *PracticeSessionStore) contentKey(accountID string) string {
	return "vocab:practice:" + accountID
}

func (s *PracticeSessionStore) progressKey(accountID string) string {
	return "vocab:practice:" + accountID + ":progress"
}

func (s *PracticeSessionStore) Load(ctx context.Context, accountID string) (vocab.PracticeSession, error) {
	data, err := s.client.Get(ctx, s.contentKey(accountID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return vocab.PracticeSession{}, shared.ErrNotFound
	}
	if err != nil {
		return vocab.PracticeSession{}, fmt.Errorf("get practice: %w", err)
	}
	var c practiceContent
	if err := json.Unmarshal(data, &c); err != nil {
		return vocab.PracticeSession{}, fmt.Errorf("unmarshal practice: %w", err)
	}
	sess := vocab.PracticeSession{
		ID:        c.ID,
		Words:     c.Words,
		Exercises: c.Exercises,
		Status:    c.Status,
		Error:     c.Error,
		Answers:   map[string]vocab.PracticeAnswer{},
	}
	if p, err := s.client.Get(ctx, s.progressKey(accountID)).Bytes(); err == nil {
		var pr practiceProgress
		if json.Unmarshal(p, &pr) == nil && pr.SessionID == c.ID {
			sess.Answers = pr.Answers
			sess.Completed = pr.Completed
		}
	}
	return sess, nil
}

func (s *PracticeSessionStore) Create(ctx context.Context, accountID string, sess vocab.PracticeSession) error {
	data, err := json.Marshal(practiceContent{
		ID:        sess.ID,
		Words:     sess.Words,
		Exercises: sess.Exercises,
		Status:    sess.Status,
		Error:     sess.Error,
	})
	if err != nil {
		return fmt.Errorf("marshal practice: %w", err)
	}
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.contentKey(accountID), data, practiceTTL)
	pipe.Del(ctx, s.progressKey(accountID))
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline: %w", err)
	}
	return nil
}

func (s *PracticeSessionStore) AppendExercises(ctx context.Context, accountID, sessionID, chunk, status, errMsg string) error {
	data, err := s.client.Get(ctx, s.contentKey(accountID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return shared.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get practice: %w", err)
	}
	var c practiceContent
	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("unmarshal practice: %w", err)
	}
	if c.ID != sessionID {
		return fmt.Errorf("%w: practice session replaced", shared.ErrNotFound)
	}
	if chunk != "" {
		if c.Exercises != "" {
			c.Exercises += "\n\n"
		}
		c.Exercises += chunk
	}
	c.Status = status
	c.Error = errMsg
	updated, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal practice: %w", err)
	}
	if err := s.client.Set(ctx, s.contentKey(accountID), updated, practiceTTL).Err(); err != nil {
		return fmt.Errorf("set practice: %w", err)
	}
	return nil
}

func (s *PracticeSessionStore) SaveProgress(ctx context.Context, accountID, sessionID string, answers map[string]vocab.PracticeAnswer, completed bool) error {
	data, err := json.Marshal(practiceProgress{SessionID: sessionID, Answers: answers, Completed: completed})
	if err != nil {
		return fmt.Errorf("marshal progress: %w", err)
	}
	if err := s.client.Set(ctx, s.progressKey(accountID), data, practiceTTL).Err(); err != nil {
		return fmt.Errorf("set progress: %w", err)
	}
	return nil
}
