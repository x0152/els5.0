package redissession

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
)

const (
	defaultPrefix    = "session:"
	accountSetSuffix = "by-account:"
)

type Store struct {
	client *redis.Client
	prefix string
}

func NewStore(client *redis.Client, prefix string) *Store {
	if prefix == "" {
		prefix = defaultPrefix
	}
	return &Store{client: client, prefix: prefix}
}

type payload struct {
	AccountID     string `json:"account_id"`
	Email         string `json:"email"`
	Role          string `json:"role,omitempty"`
	EntityID      string `json:"entity_id,omitempty"`
	IsGlobalAdmin bool   `json:"is_global_admin,omitempty"`
}

func (s *Store) Create(ctx context.Context, subject ports.SessionSubject, ttl time.Duration) (string, error) {
	if subject.IsZero() {
		return "", fmt.Errorf("session subject is empty")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("ttl must be > 0")
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	data, err := json.Marshal(payload{
		AccountID:     subject.AccountID,
		Email:         subject.Email,
		Role:          subject.Role,
		EntityID:      subject.EntityID,
		IsGlobalAdmin: subject.IsGlobalAdmin,
	})
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	tokenHash := s.hash(token)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.tokenKey(tokenHash), data, ttl)
	pipe.SAdd(ctx, s.accountKey(subject.AccountID), tokenHash)
	pipe.Expire(ctx, s.accountKey(subject.AccountID), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("redis pipeline: %w", err)
	}
	return token, nil
}

func (s *Store) Lookup(ctx context.Context, token string) (ports.SessionSubject, error) {
	if token == "" {
		return ports.SessionSubject{}, shared.ErrUnauthorized
	}
	data, err := s.client.Get(ctx, s.tokenKey(s.hash(token))).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ports.SessionSubject{}, shared.ErrUnauthorized
		}
		return ports.SessionSubject{}, fmt.Errorf("redis get: %w", err)
	}
	var pl payload
	if err := json.Unmarshal(data, &pl); err != nil {
		return ports.SessionSubject{}, fmt.Errorf("unmarshal: %w", err)
	}
	return ports.SessionSubject{
		AccountID:     pl.AccountID,
		Email:         pl.Email,
		Role:          pl.Role,
		EntityID:      pl.EntityID,
		IsGlobalAdmin: pl.IsGlobalAdmin,
	}, nil
}

func (s *Store) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	tokenHash := s.hash(token)
	data, err := s.client.Get(ctx, s.tokenKey(tokenHash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("redis get: %w", err)
	}
	var pl payload
	if err := json.Unmarshal(data, &pl); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.tokenKey(tokenHash))
	pipe.SRem(ctx, s.accountKey(pl.AccountID), tokenHash)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline: %w", err)
	}
	return nil
}

func (s *Store) RevokeByAccountID(ctx context.Context, accountID string) error {
	if accountID == "" {
		return nil
	}
	hashes, err := s.client.SMembers(ctx, s.accountKey(accountID)).Result()
	if err != nil {
		return fmt.Errorf("redis smembers: %w", err)
	}
	pipe := s.client.TxPipeline()
	for _, h := range hashes {
		pipe.Del(ctx, s.tokenKey(h))
	}
	pipe.Del(ctx, s.accountKey(accountID))
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline: %w", err)
	}
	return nil
}

func (s *Store) tokenKey(hash string) string { return s.prefix + hash }
func (s *Store) accountKey(accountID string) string {
	return s.prefix + accountSetSuffix + accountID
}

func (s *Store) hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
