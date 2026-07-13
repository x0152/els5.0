package httpx

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func Handle[In any, Out any](
	api huma.API,
	op huma.Operation,
	handle func(ctx context.Context, in *In) (Out, error),
) {
	huma.Register(api, op, func(ctx context.Context, in *In) (*Response[Out], error) {
		out, err := handle(ctx, in)
		return Return(ctx, out, err)
	})
}

func Map[In any, Out any](fn func(In) Out) func(In, error) (Out, error) {
	return func(in In, err error) (Out, error) {
		var zero Out
		if err != nil {
			return zero, err
		}
		return fn(in), nil
	}
}
