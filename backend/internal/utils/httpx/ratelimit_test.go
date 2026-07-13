package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubLimiter struct {
	allow func(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	calls int
}

func (s *stubLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	s.calls++
	return s.allow(ctx, key, limit, window)
}

func TestRateLimit_NoLimiter_PassThrough(t *testing.T) {
	mw := RateLimit(RateLimitConfig{})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRateLimit_OnlyAffectsConfiguredPaths(t *testing.T) {
	lim := &stubLimiter{allow: func(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
		return true, nil
	}}
	mw := RateLimit(RateLimitConfig{
		Limiter:      lim,
		Limit:        10,
		Window:       time.Minute,
		PathPrefixes: []string{"/api/v1/auth/"},
		Methods:      []string{http.MethodPost},
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/me", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if lim.calls != 0 {
		t.Fatalf("limiter must not be called on non-matching path, got %d", lim.calls)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if lim.calls != 0 {
		t.Fatalf("limiter must not be called on non-matching method, got %d", lim.calls)
	}
}

func TestRateLimit_DenyReturns429(t *testing.T) {
	lim := &stubLimiter{allow: func(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
		return false, nil
	}}
	mw := RateLimit(RateLimitConfig{
		Limiter:      lim,
		Limit:        1,
		Window:       time.Minute,
		PathPrefixes: []string{"/api/v1/auth/"},
		Methods:      []string{http.MethodPost},
	})
	called := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("code=%d", rec.Code)
	}
	if called {
		t.Fatal("inner handler must not be called when denied")
	}
}

func TestRateLimit_LimiterError_FailsOpen(t *testing.T) {
	lim := &stubLimiter{allow: func(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
		return false, ErrLimiterUnavailable
	}}
	mw := RateLimit(RateLimitConfig{
		Limiter:      lim,
		Limit:        1,
		Window:       time.Minute,
		PathPrefixes: []string{"/api/v1/auth/"},
		Methods:      []string{http.MethodPost},
	})
	called := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("expected fail-open to OK, got code=%d called=%v", rec.Code, called)
	}
}
