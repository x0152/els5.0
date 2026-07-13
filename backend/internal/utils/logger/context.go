package logger

import (
	"context"
	"log/slog"

	"github.com/els/backend/internal/utils/reqctx"
)

type contextHandler struct {
	base slog.Handler
}

func (h *contextHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.base.Enabled(ctx, lvl)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if rid := reqctx.RequestID(ctx); rid != "" {
		r.AddAttrs(slog.String("request_id", rid))
	}
	if u, ok := reqctx.UserOf(ctx); ok {
		if u.ID != "" {
			r.AddAttrs(slog.String("user_id", u.ID))
		}
		if u.Email != "" {
			r.AddAttrs(slog.String("user_email", u.Email))
		}
	}
	return h.base.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{base: h.base.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{base: h.base.WithGroup(name)}
}
