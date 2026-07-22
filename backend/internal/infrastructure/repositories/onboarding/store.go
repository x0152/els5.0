package onboarding

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/onboarding"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Watermarks(ctx context.Context, accountID string) (map[string]int, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT metric, value FROM onboarding_metrics WHERE account_id = $1::uuid`, accountID)
	if err != nil {
		return nil, fmt.Errorf("select onboarding metrics: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var metric string
		var value int
		if err := rows.Scan(&metric, &value); err != nil {
			return nil, fmt.Errorf("scan onboarding metric: %w", err)
		}
		out[metric] = value
	}
	return out, rows.Err()
}

func (s *Store) SaveWatermarks(ctx context.Context, accountID string, values map[string]int, now time.Time) error {
	for metric, value := range values {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO onboarding_metrics (account_id, metric, value, updated_at)
			 VALUES ($1::uuid, $2, $3, $4)
			 ON CONFLICT (account_id, metric)
			 DO UPDATE SET value = GREATEST(onboarding_metrics.value, EXCLUDED.value), updated_at = EXCLUDED.updated_at`,
			accountID, metric, value, now)
		if err != nil {
			return fmt.Errorf("upsert onboarding metric %s: %w", metric, err)
		}
	}
	return nil
}

func (s *Store) Acks(ctx context.Context, accountID string) (map[string]bool, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT item_id FROM onboarding_acks WHERE account_id = $1::uuid`, accountID)
	if err != nil {
		return nil, fmt.Errorf("select onboarding acks: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan onboarding ack: %w", err)
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (s *Store) SaveAcks(ctx context.Context, accountID string, itemIDs []string, now time.Time) error {
	for _, id := range itemIDs {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO onboarding_acks (account_id, item_id, acked_at) VALUES ($1::uuid, $2, $3)
			 ON CONFLICT (account_id, item_id) DO NOTHING`,
			accountID, id, now)
		if err != nil {
			return fmt.Errorf("insert onboarding ack %s: %w", id, err)
		}
	}
	return nil
}

func (s *Store) LiveCounts(ctx context.Context, accountID string) (map[string]int, error) {
	var workouts, vocab, chat, diary, quests, films, articles, chapters int
	// $1 is uuid, $2 the same id as text: vocab_units.account_id, quest_missions.user_id,
	// reader_books.owner_id and reader_progress.owner_id are text columns.
	err := s.pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM workout_lessons WHERE account_id = $1::uuid AND status = 'completed'),
		(SELECT count(*) FROM vocab_units WHERE account_id = $2),
		(SELECT count(*) FROM ai_messages m JOIN ai_sessions s ON s.id = m.session_id WHERE s.account_id = $1::uuid AND m.role = 'user'),
		(SELECT count(*) FROM diary_entries WHERE account_id = $1::uuid),
		(SELECT count(*) FROM quest_missions WHERE user_id = $2 AND payload->>'isComplete' = 'true'),
		(SELECT count(*) FROM film_progress WHERE owner_id = $1::uuid AND position_ms > 0),
		(SELECT count(*) FROM reader_books b JOIN reader_progress p ON p.book_id = b.id AND p.owner_id = $2
			WHERE b.owner_id = $2 AND b.kind = 'article' AND p.position > 0),
		(SELECT count(DISTINCT (kind, number)) FROM practice_progress WHERE account_id = $1::uuid AND completed)`,
		accountID, accountID).Scan(&workouts, &vocab, &chat, &diary, &quests, &films, &articles, &chapters)
	if err != nil {
		return nil, fmt.Errorf("count onboarding live metrics: %w", err)
	}
	return map[string]int{
		onboarding.MetricWorkoutsCompleted: workouts,
		onboarding.MetricVocabWords:        vocab,
		onboarding.MetricChatMessages:      chat,
		onboarding.MetricDiaryEntries:      diary,
		onboarding.MetricQuestsCompleted:   quests,
		onboarding.MetricFilmsWatched:      films,
		onboarding.MetricArticlesRead:      articles,
		onboarding.MetricBookChapters:      chapters,
	}, nil
}
