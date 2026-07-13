package settings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) ListAIProviders(ctx context.Context) ([]settings.AIProvider, error) {
	rows, err := s.pool.Query(ctx, `SELECT feature, base_url, api_key, model FROM ai_providers ORDER BY feature`)
	if err != nil {
		return nil, fmt.Errorf("list ai providers: %w", err)
	}
	defer rows.Close()
	out := make([]settings.AIProvider, 0)
	for rows.Next() {
		var p settings.AIProvider
		if err := rows.Scan(&p.Feature, &p.BaseURL, &p.APIKey, &p.Model); err != nil {
			return nil, fmt.Errorf("scan ai provider: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetAIProvider(ctx context.Context, feature settings.Feature) (settings.AIProvider, error) {
	var p settings.AIProvider
	err := s.pool.QueryRow(ctx,
		`SELECT feature, base_url, api_key, model FROM ai_providers WHERE feature = $1`, string(feature)).
		Scan(&p.Feature, &p.BaseURL, &p.APIKey, &p.Model)
	if err == pgx.ErrNoRows {
		return settings.AIProvider{}, shared.ErrNotFound
	}
	if err != nil {
		return settings.AIProvider{}, fmt.Errorf("get ai provider: %w", err)
	}
	return p, nil
}

func (s *Store) UpsertAIProvider(ctx context.Context, p settings.AIProvider) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO ai_providers (feature, base_url, api_key, model, updated_at)
		 VALUES ($1,$2,$3,$4, now())
		 ON CONFLICT (feature) DO UPDATE SET base_url=$2, api_key=$3, model=$4, updated_at=now()`,
		string(p.Feature), p.BaseURL, p.APIKey, p.Model)
	if err != nil {
		return fmt.Errorf("upsert ai provider: %w", err)
	}
	return nil
}

func (s *Store) GetFlag(ctx context.Context, key string) (bool, error) {
	var enabled bool
	err := s.pool.QueryRow(ctx, `SELECT enabled FROM platform_flags WHERE key = $1`, key).Scan(&enabled)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get flag: %w", err)
	}
	return enabled, nil
}

func (s *Store) SetFlag(ctx context.Context, key string, value bool) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO platform_flags (key, enabled, updated_at) VALUES ($1,$2, now())
		 ON CONFLICT (key) DO UPDATE SET enabled=$2, updated_at=now()`, key, value)
	if err != nil {
		return fmt.Errorf("set flag: %w", err)
	}
	return nil
}
