package httpx

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/els/backend/internal/utils/reqctx"
)

func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ow := newObservedWriter(w)
			next.ServeHTTP(ow, r)
			if reqctx.IsSilent(r.Context()) {
				return
			}
			latency := time.Since(start)

			level := slog.LevelInfo
			outcome := "success"
			switch {
			case ow.Status() >= 500:
				level = slog.LevelError
				outcome = "server_error"
			case ow.Status() >= 400:
				outcome = "client_error"
			}

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ow.Status()),
				slog.String("outcome", outcome),
				slog.Int64("latency_ms", latency.Milliseconds()),
				slog.Int64("bytes_out", ow.Bytes()),
				slog.String("remote_ip", clientIP(r)),
			}
			if ua := r.UserAgent(); ua != "" {
				attrs = append(attrs, slog.String("user_agent", ua))
			}

			logger.LogAttrs(r.Context(), level, "http request", attrs...)
		})
	}
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			v = v[:i]
		}
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return strings.TrimSpace(v)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}
