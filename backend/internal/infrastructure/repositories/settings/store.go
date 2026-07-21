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
	rows, err := s.pool.Query(ctx, `SELECT feature, kind, base_url, api_key, model, params FROM ai_providers ORDER BY feature`)
	if err != nil {
		return nil, fmt.Errorf("list ai providers: %w", err)
	}
	defer rows.Close()
	out := make([]settings.AIProvider, 0)
	for rows.Next() {
		var p settings.AIProvider
		if err := rows.Scan(&p.Feature, &p.Kind, &p.BaseURL, &p.APIKey, &p.Model, &p.Params); err != nil {
			return nil, fmt.Errorf("scan ai provider: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetAIProvider(ctx context.Context, feature settings.Feature) (settings.AIProvider, error) {
	var p settings.AIProvider
	err := s.pool.QueryRow(ctx,
		`SELECT feature, kind, base_url, api_key, model, params FROM ai_providers WHERE feature = $1`, string(feature)).
		Scan(&p.Feature, &p.Kind, &p.BaseURL, &p.APIKey, &p.Model, &p.Params)
	if err == pgx.ErrNoRows {
		return settings.AIProvider{}, shared.ErrNotFound
	}
	if err != nil {
		return settings.AIProvider{}, fmt.Errorf("get ai provider: %w", err)
	}
	return p, nil
}

func (s *Store) UpsertAIProvider(ctx context.Context, p settings.AIProvider) error {
	if p.Kind == "" {
		p.Kind = settings.KindOpenAI
	}
	if p.Params == nil {
		p.Params = map[string]string{}
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO ai_providers (feature, kind, base_url, api_key, model, params, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6, now())
		 ON CONFLICT (feature) DO UPDATE SET kind=$2, base_url=$3, api_key=$4, model=$5, params=$6, updated_at=now()`,
		string(p.Feature), string(p.Kind), p.BaseURL, p.APIKey, p.Model, p.Params)
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
