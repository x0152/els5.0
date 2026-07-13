package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/infrastructure/postgres/sqlc"
	"github.com/els/backend/internal/utils/database"
)

type txCtxKey struct{}

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx)
	return tx, ok
}

func QueriesFromContext(ctx context.Context, base *sqlc.Queries) *sqlc.Queries {
	if tx, ok := TxFromContext(ctx); ok {
		return base.WithTx(tx)
	}
	return base
}

func RunTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	if _, ok := TxFromContext(ctx); ok {
		return fn(ctx)
	}
	return InTx(ctx, pool, func(tx pgx.Tx) error {
		return fn(WithTx(ctx, tx))
	})
}

type pgxTxRunner struct {
	pool *pgxpool.Pool
}

func NewTxRunner(pool *pgxpool.Pool) database.TxRunner {
	return &pgxTxRunner{pool: pool}
}

func (r *pgxTxRunner) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return RunTx(ctx, r.pool, fn)
}
