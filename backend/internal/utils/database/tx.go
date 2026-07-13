package database

import "context"

type TxRunner interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

type TxRunnerFunc func(ctx context.Context, fn func(ctx context.Context) error) error

func (f TxRunnerFunc) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return f(ctx, fn)
}

func Noop() TxRunner {
	return TxRunnerFunc(func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	})
}
