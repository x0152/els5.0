package rediscode

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
	defaultPrefix    = "invite:"
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
	Purpose   string `json:"purpose"`
	AccountID string `json:"account_id"`
	Reusable  bool   `json:"reusable,omitempty"`
}

func (s *Store) Issue(ctx context.Context, tok ports.InviteToken, ttl time.Duration) (string, error) {
	if tok.Purpose == "" {
		return "", fmt.Errorf("invite purpose must not be empty")
	}
	if tok.AccountID == "" {
		return "", fmt.Errorf("invite account_id must not be empty")
	}
	if ttl < 0 {
		return "", fmt.Errorf("ttl must be >= 0")
	}
	if err := s.revokeByAccountPurpose(ctx, tok.AccountID, string(tok.Purpose)); err != nil {
		return "", err
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	data, err := json.Marshal(payload{Purpose: string(tok.Purpose), AccountID: tok.AccountID, Reusable: tok.Reusable})
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	tokenHash := s.hash(token)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.tokenKey(tokenHash), data, ttl)
	pipe.SAdd(ctx, s.accountKey(tok.AccountID, string(tok.Purpose)), tokenHash)
	if ttl > 0 {
		pipe.Expire(ctx, s.accountKey(tok.AccountID, string(tok.Purpose)), ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("redis pipeline: %w", err)
	}
	return token, nil
}

func (s *Store) Consume(ctx context.Context, token string) (ports.InviteToken, error) {
	if token == "" {
		return ports.InviteToken{}, shared.ErrUnauthorized
	}
	tokenHash := s.hash(token)
	data, err := s.client.GetDel(ctx, s.tokenKey(tokenHash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ports.InviteToken{}, shared.ErrUnauthorized
		}
		return ports.InviteToken{}, fmt.Errorf("redis getdel: %w", err)
	}
	var pl payload
	if err := json.Unmarshal(data, &pl); err != nil {
		return ports.InviteToken{}, fmt.Errorf("unmarshal: %w", err)
	}
	if pl.Reusable {
		if err := s.client.Set(ctx, s.tokenKey(tokenHash), data, 0).Err(); err != nil {
			return ports.InviteToken{}, fmt.Errorf("redis restore: %w", err)
		}
	} else if err := s.client.SRem(ctx, s.accountKey(pl.AccountID, pl.Purpose), tokenHash).Err(); err != nil {
		return ports.InviteToken{}, fmt.Errorf("redis srem: %w", err)
	}
	return ports.InviteToken{
		Purpose:   ports.InviteTokenPurpose(pl.Purpose),
		AccountID: pl.AccountID,
		Reusable:  pl.Reusable,
	}, nil
}

func (s *Store) revokeByAccountPurpose(ctx context.Context, accountID, purpose string) error {
	setKey := s.accountKey(accountID, purpose)
	hashes, err := s.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return fmt.Errorf("redis smembers: %w", err)
	}
	if len(hashes) == 0 {
		return nil
	}
	pipe := s.client.TxPipeline()
	for _, h := range hashes {
		pipe.Del(ctx, s.tokenKey(h))
	}
	pipe.Del(ctx, setKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline: %w", err)
	}
	return nil
}

func (s *Store) tokenKey(hash string) string { return s.prefix + hash }
func (s *Store) accountKey(accountID, purpose string) string {
	return s.prefix + accountSetSuffix + purpose + ":" + accountID
}

func (s *Store) hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
