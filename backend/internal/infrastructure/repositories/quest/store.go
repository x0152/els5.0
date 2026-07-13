package quest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Insert(ctx context.Context, userID string, mission *quest.CustomMission) error {
	return s.insert(ctx, userID, userID, mission)
}

func (s *Store) Fork(ctx context.Context, userID, authorID string, mission *quest.CustomMission) error {
	return s.insert(ctx, userID, authorID, mission)
}

func (s *Store) insert(ctx context.Context, userID, authorID string, mission *quest.CustomMission) error {
	payload, err := json.Marshal(mission)
	if err != nil {
		return fmt.Errorf("marshal mission: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO quest_missions (id, user_id, author_id, status, error, payload, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,now(),now())`,
		mission.ID, userID, authorID, status(mission), mission.GenerationError, payload); err != nil {
		return fmt.Errorf("insert mission: %w", err)
	}
	return nil
}

func (s *Store) Save(ctx context.Context, userID string, mission *quest.CustomMission) error {
	payload, err := json.Marshal(mission)
	if err != nil {
		return fmt.Errorf("marshal mission: %w", err)
	}
	ct, err := s.pool.Exec(ctx, `UPDATE quest_missions SET status = $1, error = $2, payload = $3, updated_at = now()
		WHERE id = $4 AND user_id = $5`,
		status(mission), mission.GenerationError, payload, mission.ID, userID)
	if err != nil {
		return fmt.Errorf("save mission: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// Update performs read-modify-write atomically: competing goroutines
// (dialogue, summarization, image generation) do not overwrite each other's changes.
func (s *Store) Update(ctx context.Context, userID, id string, mutate func(*quest.CustomMission) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var payload []byte
	err = tx.QueryRow(ctx, `SELECT payload FROM quest_missions WHERE id = $1 AND user_id = $2 FOR UPDATE`,
		id, userID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock mission: %w", err)
	}
	mission, err := unmarshalMission(payload)
	if err != nil {
		return err
	}
	if err := mutate(mission); err != nil {
		return err
	}
	updated, err := json.Marshal(mission)
	if err != nil {
		return fmt.Errorf("marshal mission: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE quest_missions SET status = $1, error = $2, payload = $3, updated_at = now()
		WHERE id = $4 AND user_id = $5`,
		status(mission), mission.GenerationError, updated, id, userID); err != nil {
		return fmt.Errorf("update mission: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Store) GetByID(ctx context.Context, userID, id string) (*quest.CustomMission, error) {
	var payload []byte
	err := s.pool.QueryRow(ctx, `SELECT payload FROM quest_missions WHERE id = $1 AND user_id = $2`, id, userID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get mission: %w", err)
	}
	return unmarshalMission(payload)
}

func (s *Store) GetOrigin(ctx context.Context, id string) (*quest.CustomMission, string, error) {
	var payload []byte
	var authorID string
	err := s.pool.QueryRow(ctx, `SELECT payload, author_id FROM quest_missions WHERE id = $1 AND user_id = author_id`, id).
		Scan(&payload, &authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", shared.ErrNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("get origin mission: %w", err)
	}
	m, err := unmarshalMission(payload)
	if err != nil {
		return nil, "", err
	}
	return m, authorID, nil
}

func (s *Store) GetAllByID(ctx context.Context, id string) ([]*quest.CustomMission, error) {
	rows, err := s.pool.Query(ctx, `SELECT payload FROM quest_missions WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("list mission copies: %w", err)
	}
	defer rows.Close()
	out := []*quest.CustomMission{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan mission: %w", err)
		}
		m, err := unmarshalMission(payload)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) List(ctx context.Context, userID string) ([]quest.MissionCatalogItem, error) {
	rows, err := s.pool.Query(ctx, `SELECT COALESCE(u.payload, o.payload), u.payload IS NOT NULL
		FROM quest_missions o
		-- catalog: all originals, overlaid with the viewer's personal copy if they already played
		LEFT JOIN quest_missions u ON u.id = o.id AND u.user_id = $1
		WHERE o.user_id = o.author_id
		ORDER BY o.created_at DESC LIMIT 500`, userID)
	if err != nil {
		return nil, fmt.Errorf("list missions: %w", err)
	}
	defer rows.Close()
	out := []quest.MissionCatalogItem{}
	for rows.Next() {
		var payload []byte
		var started bool
		if err := rows.Scan(&payload, &started); err != nil {
			return nil, fmt.Errorf("scan mission: %w", err)
		}
		m, err := unmarshalMission(payload)
		if err != nil {
			return nil, err
		}
		out = append(out, quest.MissionCatalogItem{Mission: m, Started: started})
	}
	return out, rows.Err()
}

// FailStaleGenerating marks missions and their images left mid-generation
// (e.g. after a restart) as failed, so the UI stops showing eternal spinners.
func (s *Store) FailStaleGenerating(ctx context.Context) error {
	rows, err := s.pool.Query(ctx, `SELECT user_id, payload FROM quest_missions
		WHERE status = $1
		   OR payload->>'coverImageStatus' = $1
		   -- generating may linger in scene and avatar statuses (map[string]string)
		   OR EXISTS (SELECT 1 FROM jsonb_each_text(COALESCE(payload->'sceneImageStatus', '{}'::jsonb)) s WHERE s.value = $1)
		   OR EXISTS (SELECT 1 FROM jsonb_each_text(COALESCE(payload->'characterAvatarStatus', '{}'::jsonb)) a WHERE a.value = $1)`,
		quest.GenerationStatusGenerating)
	if err != nil {
		return fmt.Errorf("list stale missions: %w", err)
	}
	type record struct {
		userID  string
		mission *quest.CustomMission
	}
	var stale []record
	for rows.Next() {
		var userID string
		var payload []byte
		if err := rows.Scan(&userID, &payload); err != nil {
			rows.Close()
			return fmt.Errorf("scan stale mission: %w", err)
		}
		m, err := unmarshalMission(payload)
		if err != nil {
			rows.Close()
			return err
		}
		stale = append(stale, record{userID: userID, mission: m})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, r := range stale {
		if failStaleMission(r.mission) {
			if err := s.Save(ctx, r.userID, r.mission); err != nil {
				return err
			}
		}
	}
	return nil
}

func failStaleMission(m *quest.CustomMission) bool {
	const interrupted = "generation interrupted"
	changed := false
	if m.GenerationStatus == quest.GenerationStatusGenerating {
		m.GenerationStatus = quest.GenerationStatusError
		m.GenerationError = interrupted
		changed = true
	}
	if m.CoverImageStatus == quest.GenerationStatusGenerating {
		m.CoverImageStatus = quest.GenerationStatusError
		if m.CoverImageError == "" {
			m.CoverImageError = interrupted
		}
		changed = true
	}
	if failStaleStatusMap(m.SceneImageStatus, &m.SceneImageErrors, interrupted) {
		changed = true
	}
	if failStaleStatusMap(m.CharacterAvatarStatus, &m.CharacterAvatarErrors, interrupted) {
		changed = true
	}
	return changed
}

func failStaleStatusMap(statuses map[string]string, errs *map[string]string, msg string) bool {
	changed := false
	for key, st := range statuses {
		if st == quest.GenerationStatusGenerating {
			statuses[key] = quest.GenerationStatusError
			if *errs == nil {
				*errs = map[string]string{}
			}
			if (*errs)[key] == "" {
				(*errs)[key] = msg
			}
			changed = true
		}
	}
	return changed
}

func (s *Store) Delete(ctx context.Context, id string) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM quest_missions WHERE id = $1 AND user_id = author_id`, id)
	if err != nil {
		return fmt.Errorf("delete mission: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

type ProfileStore struct {
	pool *pgxpool.Pool
}

func NewProfileStore(pool *pgxpool.Pool) *ProfileStore { return &ProfileStore{pool: pool} }

func (s *ProfileStore) Get(ctx context.Context, userID string) (quest.PlayerProfile, error) {
	var payload []byte
	err := s.pool.QueryRow(ctx, `SELECT payload FROM quest_profiles WHERE user_id = $1`, userID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return quest.PlayerProfile{}, shared.ErrNotFound
	}
	if err != nil {
		return quest.PlayerProfile{}, fmt.Errorf("get profile: %w", err)
	}
	var profile quest.PlayerProfile
	if err := json.Unmarshal(payload, &profile); err != nil {
		return quest.PlayerProfile{}, fmt.Errorf("unmarshal profile: %w", err)
	}
	return profile, nil
}

func (s *ProfileStore) Save(ctx context.Context, userID string, profile quest.PlayerProfile) error {
	payload, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO quest_profiles (user_id, payload, created_at, updated_at)
		VALUES ($1,$2,now(),now())
		ON CONFLICT (user_id) DO UPDATE SET payload = EXCLUDED.payload, updated_at = now()`,
		userID, payload); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

func status(mission *quest.CustomMission) string {
	if mission.GenerationStatus == "" {
		return quest.GenerationStatusGenerating
	}
	return mission.GenerationStatus
}

func unmarshalMission(payload []byte) (*quest.CustomMission, error) {
	var m quest.CustomMission
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil, fmt.Errorf("unmarshal mission: %w", err)
	}
	return &m, nil
}
