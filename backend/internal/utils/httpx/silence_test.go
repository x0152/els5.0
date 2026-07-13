package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/reqctx"
)

func TestSilenceProbes_MarksHealthAndReady(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path       string
		wantSilent bool
	}{
		{"/health", true},
		{"/ready", true},
		{"/v1/experts", false},
		{"/healthz", false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			var gotSilent bool
			h := SilenceProbes(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotSilent = reqctx.IsSilent(r.Context())
				w.WriteHeader(http.StatusOK)
			}))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if gotSilent != tc.wantSilent {
				t.Fatalf("path=%s: want silent=%v, got %v", tc.path, tc.wantSilent, gotSilent)
			}
		})
	}
}

func TestSilencePaths_Custom(t *testing.T) {
	t.Parallel()
	mw := SilencePaths("/metrics", "/debug/vars")

	var gotSilent bool
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSilent = reqctx.IsSilent(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !gotSilent {
		t.Fatal("custom silence middleware should mark /metrics as silent")
	}

	gotSilent = false
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/foo", nil))
	if gotSilent {
		t.Fatal("custom silence middleware must not mark unrelated paths")
	}
}

func TestAccessLog_SkipsSilentRequests(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	chain := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}),
		RequestID,
		SilenceProbes,
		AccessLog(logger),
	)

	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ready", nil))

	if buf.Len() != 0 {
		t.Fatalf("expected no access-log for probes, got: %s", buf.String())
	}

	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/v1/foo", nil))

	if !strings.Contains(buf.String(), `"path":"/v1/foo"`) {
		t.Fatalf("expected access-log for /v1/foo, got: %s", buf.String())
	}
}

func TestLogResponseError_SkipsSilentContext(t *testing.T) {
	t.Parallel()
	prev := slog.Default()
	defer slog.SetDefault(prev)

	buf := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ctx := reqctx.WithSilent(context.Background())
	LogResponseError(ctx, errors.New("probe failure"), http.StatusServiceUnavailable, CodeUnavailable)
	LogResponseError(ctx, shared.ErrUnavailable, http.StatusServiceUnavailable, CodeUnavailable)

	if buf.Len() != 0 {
		t.Fatalf("expected no error-log for silent ctx, got: %s", buf.String())
	}
}

func TestAccessLog_SilenceDoesNotLeakAcrossRequests(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	chain := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantSilent := r.URL.Path == "/health"
		if got := reqctx.IsSilent(r.Context()); got != wantSilent {
			t.Fatalf("path=%s want silent=%v, got %v", r.URL.Path, wantSilent, got)
		}
	}),
		RequestID,
		SilenceProbes,
		AccessLog(logger),
	)

	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/v1/foo", nil))
}
