package httpx

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"

	"github.com/els/backend/internal/utils/reqctx"
)

func LogResponseError(ctx context.Context, err error, status int, code ErrorCode) {
	if err == nil || status < 400 {
		return
	}
	if reqctx.IsSilent(ctx) {
		return
	}
	if isOpaqueInternalEnvelope(err) {
		return
	}

	level := slog.LevelWarn
	if status >= 500 {
		level = slog.LevelError
	}

	attrs := []slog.Attr{
		slog.String("err", err.Error()),
		slog.Int("status", status),
		slog.String("code", string(code)),
	}
	attrs = appendErrorTypeAttr(attrs, err)

	if level == slog.LevelError {
		attrs = append(attrs, slog.String("stack", stackFor(err)))
	}

	slog.LogAttrs(ctx, level, "request error", attrs...)
}

func isOpaqueInternalEnvelope(err error) bool {
	var he *Error
	if !errors.As(err, &he) {
		return false
	}
	return he.Status >= 500 && he.Err == nil && len(he.Stack) == 0
}

func stackFor(err error) string {
	var he *Error
	if errors.As(err, &he) && len(he.Stack) > 0 {
		return string(he.Stack)
	}
	return string(debug.Stack())
}

func appendErrorTypeAttr(attrs []slog.Attr, err error) []slog.Attr {
	var he *Error
	if errors.As(err, &he) {
		attrs = append(attrs, slog.String("err_type", "httpx.Error"))
	}
	return attrs
}
