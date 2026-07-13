package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/core"
	"github.com/els/backend/internal/domain/shared/vo"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const rawColumns = `id, user_id, client_id, status, skill, text, target, outcome, context, source, meta,
	occurred_at, created_at, processed_at, error`

const eventColumns = `id, raw_event_id, user_id, client_id, type, action, lemma, pos, grammar_key, outcome,
	error, context, source, meta, occurred_at, created_at`

func (s *Store) InsertRaw(ctx context.Context, e core.RawEvent) error {
	return insertRaw(ctx, s.pool, e)
}

func (s *Store) InsertRawBatch(ctx context.Context, events []core.RawEvent) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, e := range events {
		if err := insertRaw(ctx, tx, e); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func insertRaw(ctx context.Context, db execer, e core.RawEvent) error {
	_, err := db.Exec(ctx, `INSERT INTO raw_events (`+rawColumns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		e.ID, e.UserID, e.ClientID, e.Status, e.Skill, e.Text, e.Target, e.Outcome, e.Context,
		jsonbOrEmpty(e.Source), jsonbOrEmpty(e.Meta), e.OccurredAt, e.CreatedAt, e.ProcessedAt, e.Error)
	if err != nil {
		return fmt.Errorf("insert raw event: %w", err)
	}
	return nil
}

func (s *Store) ListRaw(ctx context.Context, userID, status string) ([]core.RawEvent, error) {
	statuses := []string{status}
	if status == string(core.StatusPending) {
		statuses = append(statuses, string(core.StatusProcessing))
	}
	rows, err := s.pool.Query(ctx, `SELECT `+rawColumns+` FROM raw_events
		WHERE user_id = $1 AND status = ANY($2) ORDER BY created_at DESC LIMIT 500`, userID, statuses)
	if err != nil {
		return nil, fmt.Errorf("list raw events: %w", err)
	}
	defer rows.Close()
	out := []core.RawEvent{}
	for rows.Next() {
		var e core.RawEvent
		var source, meta []byte
		if err := rows.Scan(&e.ID, &e.UserID, &e.ClientID, &e.Status, &e.Skill, &e.Text, &e.Target, &e.Outcome,
			&e.Context, &source, &meta, &e.OccurredAt, &e.CreatedAt, &e.ProcessedAt, &e.Error); err != nil {
			return nil, fmt.Errorf("scan raw event: %w", err)
		}
		e.Source, e.Meta = source, meta
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) FindRawByText(ctx context.Context, userID, skill, text string) (core.RawEvent, bool, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+rawColumns+` FROM raw_events
		WHERE user_id = $1 AND skill = $2 AND text = $3 ORDER BY created_at DESC LIMIT 1`, userID, skill, text)
	if err != nil {
		return core.RawEvent{}, false, fmt.Errorf("find raw by text: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return core.RawEvent{}, false, rows.Err()
	}
	var e core.RawEvent
	var source, meta []byte
	if err := rows.Scan(&e.ID, &e.UserID, &e.ClientID, &e.Status, &e.Skill, &e.Text, &e.Target, &e.Outcome,
		&e.Context, &source, &meta, &e.OccurredAt, &e.CreatedAt, &e.ProcessedAt, &e.Error); err != nil {
		return core.RawEvent{}, false, fmt.Errorf("scan raw event: %w", err)
	}
	e.Source, e.Meta = source, meta
	return e, true, nil
}

func (s *Store) SetRawOutcome(ctx context.Context, rawID, outcome string) error {
	_, err := s.pool.Exec(ctx, `UPDATE raw_events SET outcome = $1 WHERE id = $2`, outcome, rawID)
	return err
}

func (s *Store) ListAllRaw(ctx context.Context, userID string) ([]core.RawEvent, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+rawColumns+` FROM raw_events
		WHERE user_id = $1 ORDER BY occurred_at DESC LIMIT 500`, userID)
	if err != nil {
		return nil, fmt.Errorf("list all raw events: %w", err)
	}
	defer rows.Close()
	out := []core.RawEvent{}
	for rows.Next() {
		var e core.RawEvent
		var source, meta []byte
		if err := rows.Scan(&e.ID, &e.UserID, &e.ClientID, &e.Status, &e.Skill, &e.Text, &e.Target, &e.Outcome,
			&e.Context, &source, &meta, &e.OccurredAt, &e.CreatedAt, &e.ProcessedAt, &e.Error); err != nil {
			return nil, fmt.Errorf("scan raw event: %w", err)
		}
		e.Source, e.Meta = source, meta
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) ClaimPendingRaw(ctx context.Context, limit int) ([]core.RawEvent, error) {
	return s.claimByStatus(ctx, string(core.StatusPending), limit)
}

func (s *Store) ClaimFailedRaw(ctx context.Context, limit int) ([]core.RawEvent, error) {
	return s.claimByStatus(ctx, string(core.StatusFailed), limit)
}

func (s *Store) claimByStatus(ctx context.Context, status string, limit int) ([]core.RawEvent, error) {
	rows, err := s.pool.Query(ctx, `UPDATE raw_events SET status = 'processing', claimed_at = now()
		WHERE id IN (
			-- SKIP LOCKED: parallel instances do not claim the same events
			SELECT id FROM raw_events WHERE status = $1 ORDER BY created_at LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		RETURNING `+rawColumns, status, limit)
	if err != nil {
		return nil, fmt.Errorf("claim raw by status: %w", err)
	}
	defer rows.Close()
	out := []core.RawEvent{}
	for rows.Next() {
		var e core.RawEvent
		var source, meta []byte
		if err := rows.Scan(&e.ID, &e.UserID, &e.ClientID, &e.Status, &e.Skill, &e.Text, &e.Target, &e.Outcome,
			&e.Context, &source, &meta, &e.OccurredAt, &e.CreatedAt, &e.ProcessedAt, &e.Error); err != nil {
			return nil, fmt.Errorf("scan raw event: %w", err)
		}
		e.Source, e.Meta = source, meta
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) ListEvents(ctx context.Context, userID string) ([]core.Event, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+eventColumns+` FROM events
		WHERE user_id = $1 ORDER BY created_at DESC LIMIT 500`, userID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	out := []core.Event{}
	for rows.Next() {
		var e core.Event
		var errJSON, source, meta []byte
		if err := rows.Scan(&e.ID, &e.RawEventID, &e.UserID, &e.ClientID, &e.Type, &e.Action, &e.Lemma, &e.POS,
			&e.GrammarKey, &e.Outcome, &errJSON, &e.Context, &source, &meta, &e.OccurredAt, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Source, e.Meta = source, meta
		if len(errJSON) > 0 {
			var ee core.EventError
			if json.Unmarshal(errJSON, &ee) == nil {
				e.Error = &ee
			}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

const wordColumns = `id, key, lemma, pos, type, cefr, frequency, enriched, is_stop, created_at, updated_at`

const grammarColumns = `id, key, parent_key, title, cefr_level, enriched, created_at, updated_at`

func (s *Store) ListWords(ctx context.Context, limit int) ([]core.Word, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+wordColumns+` FROM words ORDER BY updated_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list words: %w", err)
	}
	return scanWords(rows)
}

func (s *Store) ListUnenrichedWords(ctx context.Context, limit int) ([]core.Word, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+wordColumns+` FROM words WHERE enriched = false ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list unenriched words: %w", err)
	}
	return scanWords(rows)
}

func (s *Store) ListGrammarRules(ctx context.Context, limit int) ([]core.GrammarRule, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+grammarColumns+` FROM grammar_rules ORDER BY updated_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list grammar rules: %w", err)
	}
	return scanGrammarRules(rows)
}

func (s *Store) ListUnenrichedGrammarRules(ctx context.Context, limit int) ([]core.GrammarRule, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+grammarColumns+` FROM grammar_rules WHERE enriched = false ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list unenriched grammar rules: %w", err)
	}
	return scanGrammarRules(rows)
}

func (s *Store) EnrichWord(ctx context.Context, id, cefr string, frequency float64) error {
	_, err := s.pool.Exec(ctx, `UPDATE words SET cefr = $1, frequency = $2, enriched = true, updated_at = now() WHERE id = $3`,
		cefr, frequency, id)
	return err
}

func (s *Store) EnrichGrammarRule(ctx context.Context, id, title, cefrLevel string) error {
	_, err := s.pool.Exec(ctx, `UPDATE grammar_rules SET title = $1, cefr_level = $2, enriched = true, updated_at = now() WHERE id = $3`,
		title, cefrLevel, id)
	return err
}

func scanWords(rows pgx.Rows) ([]core.Word, error) {
	defer rows.Close()
	out := []core.Word{}
	for rows.Next() {
		var w core.Word
		var freq *float64
		if err := rows.Scan(&w.ID, &w.Key, &w.Lemma, &w.POS, &w.Type, &w.CEFR, &freq,
			&w.Enriched, &w.IsStop, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan word: %w", err)
		}
		if freq != nil {
			w.Frequency = *freq
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func scanGrammarRules(rows pgx.Rows) ([]core.GrammarRule, error) {
	defer rows.Close()
	out := []core.GrammarRule{}
	for rows.Next() {
		var g core.GrammarRule
		if err := rows.Scan(&g.ID, &g.Key, &g.ParentKey, &g.Title, &g.CEFRLevel,
			&g.Enriched, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan grammar rule: %w", err)
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) Complete(ctx context.Context, rawID string, events []core.Event) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, e := range events {
		var wordID, ruleID any
		if key := core.WordKey(e.Lemma, e.POS); key != "" {
			id, err := upsertWord(ctx, tx, key, e.Lemma, e.POS, e.Unit)
			if err != nil {
				return fmt.Errorf("upsert word: %w", err)
			}
			wordID = id
		}
		if gk := strings.TrimSpace(e.GrammarKey); gk != "" {
			id, err := upsertGrammarRule(ctx, tx, gk)
			if err != nil {
				return fmt.Errorf("upsert grammar rule: %w", err)
			}
			ruleID = id
		}
		if _, err := tx.Exec(ctx, `INSERT INTO events (`+eventColumns+`, word_id, grammar_rule_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
			e.ID, rawID, e.UserID, e.ClientID, e.Type, e.Action, e.Lemma, e.POS, e.GrammarKey, e.Outcome,
			marshalError(e.Error), e.Context, jsonbOrEmpty(e.Source), jsonbOrEmpty(e.Meta), e.OccurredAt, e.CreatedAt, wordID, ruleID); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE raw_events SET status = $1, processed_at = now() WHERE id = $2`,
		string(core.StatusProcessed), rawID); err != nil {
		return fmt.Errorf("complete raw: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Store) DeleteRows(ctx context.Context, userID, kind string, ids []string) (int64, error) {
	switch kind {
	case "events":
		ct, err := s.pool.Exec(ctx, `DELETE FROM events WHERE user_id = $1 AND id = ANY($2)`, userID, ids)
		if err != nil {
			return 0, fmt.Errorf("delete events: %w", err)
		}
		return ct.RowsAffected(), nil
	case "raw":
		ct, err := s.pool.Exec(ctx, `DELETE FROM raw_events WHERE user_id = $1 AND id = ANY($2)`, userID, ids)
		if err != nil {
			return 0, fmt.Errorf("delete raw events: %w", err)
		}
		return ct.RowsAffected(), nil
	case "words":
		return s.deleteCatalog(ctx, ids, "word_id", "words")
	case "rules":
		return s.deleteCatalog(ctx, ids, "grammar_rule_id", "grammar_rules")
	default:
		return 0, fmt.Errorf("delete rows: unknown kind %q", kind)
	}
}

func (s *Store) deleteCatalog(ctx context.Context, ids []string, fk, table string) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE events SET `+fk+` = NULL WHERE `+fk+` = ANY($1)`, ids); err != nil {
		return 0, fmt.Errorf("detach %s: %w", table, err)
	}
	ct, err := tx.Exec(ctx, `DELETE FROM `+table+` WHERE id = ANY($1)`, ids)
	if err != nil {
		return 0, fmt.Errorf("delete %s: %w", table, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

func (s *Store) WipeUser(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, q := range []string{
		`DELETE FROM events WHERE user_id = $1`,
		`DELETE FROM raw_events WHERE user_id = $1`,
	} {
		if _, err := tx.Exec(ctx, q, userID); err != nil {
			return fmt.Errorf("wipe user: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) RequeueStaleProcessing(ctx context.Context, olderThan time.Duration) error {
	_, err := s.pool.Exec(ctx, `UPDATE raw_events SET status = 'pending', claimed_at = NULL
		WHERE status = 'processing' AND claimed_at < now() - $1::interval`,
		fmt.Sprintf("%d seconds", int(olderThan.Seconds())))
	if err != nil {
		return fmt.Errorf("requeue stale processing: %w", err)
	}
	return nil
}

func (s *Store) Fail(ctx context.Context, rawID, reason string) error {
	_, err := s.pool.Exec(ctx, `UPDATE raw_events SET status = $1, processed_at = now(), error = $2 WHERE id = $3`,
		string(core.StatusFailed), reason, rawID)
	return err
}

func upsertWord(ctx context.Context, tx pgx.Tx, key, lemma, pos, unit string) (string, error) {
	if unit == "" {
		unit = "word"
	}
	var id string
	err := tx.QueryRow(ctx, `INSERT INTO words (id, key, lemma, pos, type, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,now(),now())
		ON CONFLICT (key) DO UPDATE SET updated_at = now(),
			type = CASE WHEN words.type IS NULL OR words.type = '' THEN EXCLUDED.type ELSE words.type END
		RETURNING id`, vo.NewID().String(), key, lemma, pos, unit).Scan(&id)
	return id, err
}

func upsertGrammarRule(ctx context.Context, tx pgx.Tx, key string) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `INSERT INTO grammar_rules (id, key, parent_key, created_at, updated_at)
		VALUES ($1,$2,$3,now(),now())
		ON CONFLICT (key) DO UPDATE SET updated_at = now()
		RETURNING id`, vo.NewID().String(), key, core.GrammarParentKey(key)).Scan(&id)
	return id, err
}

func jsonbOrEmpty(b json.RawMessage) string {
	if len(b) == 0 {
		return "{}"
	}
	return string(b)
}

func marshalError(e *core.EventError) any {
	if e == nil {
		return nil
	}
	b, _ := json.Marshal(e)
	return string(b)
}
