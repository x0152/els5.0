package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type RateLimitConfig struct {
	Limiter      RateLimiter
	Limit        int
	Window       time.Duration
	PathPrefixes []string
	Methods      []string
	KeyPrefix    string
	Logger       *slog.Logger
}

func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.Limiter == nil || cfg.Limit <= 0 || cfg.Window <= 0 || len(cfg.PathPrefixes) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	methods := make(map[string]struct{}, len(cfg.Methods))
	for _, m := range cfg.Methods {
		methods[strings.ToUpper(m)] = struct{}{}
	}
	if len(methods) == 0 {
		methods[http.MethodPost] = struct{}{}
	}
	prefix := strings.TrimSuffix(cfg.KeyPrefix, ":")
	if prefix == "" {
		prefix = "ratelimit"
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := methods[r.Method]; !ok {
				next.ServeHTTP(w, r)
				return
			}
			if !pathMatchesAny(r.URL.Path, cfg.PathPrefixes) {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			key := prefix + ":ip:" + ip + ":" + r.URL.Path
			ok, err := cfg.Limiter.Allow(r.Context(), key, cfg.Limit, cfg.Window)
			if err != nil {
				logger.WarnContext(r.Context(), "rate limit check failed",
					slog.String("err", err.Error()),
					slog.String("path", r.URL.Path),
				)
				next.ServeHTTP(w, r)
				return
			}
			if !ok {
				WriteError(w, r, NewError(http.StatusTooManyRequests, CodeTooManyRequests, "too many requests"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func pathMatchesAny(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

var ErrLimiterUnavailable = errors.New("rate limiter unavailable")
