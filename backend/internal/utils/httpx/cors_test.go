package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCORS_Disabled_NoOrigins(t *testing.T) {
	mw := CORS(CORSConfig{})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS headers, got %q", got)
	}
}

func TestCORS_AllowedOrigin_PassesThrough(t *testing.T) {
	mw := CORS(CORSConfig{AllowedOrigins: []string{"https://app.example"}, MaxAge: 600 * time.Second})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("Allow-Origin=%q", got)
	}
	if rec.Header().Get("Vary") == "" {
		t.Fatalf("expected Vary header")
	}
}

func TestCORS_DisallowedOrigin_NoHeaders(t *testing.T) {
	mw := CORS(CORSConfig{AllowedOrigins: []string{"https://app.example"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Allow-Origin for disallowed, got %q", got)
	}
}

func TestCORS_PreflightAllowed(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowedOrigins: []string{"https://app.example"},
		AllowedMethods: []string{http.MethodPost, http.MethodGet},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		ExposedHeaders: []string{"X-Request-ID"},
		MaxAge:         5 * time.Minute,
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight must not call inner handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight code=%d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("missing Allow-Methods")
	}
	if rec.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatalf("missing Allow-Headers")
	}
	if rec.Header().Get("Access-Control-Max-Age") != "300" {
		t.Fatalf("Max-Age=%q", rec.Header().Get("Access-Control-Max-Age"))
	}
	if rec.Header().Get("Access-Control-Expose-Headers") != "X-Request-ID" {
		t.Fatalf("Expose=%q", rec.Header().Get("Access-Control-Expose-Headers"))
	}
}

func TestCORS_PreflightDisallowed_403(t *testing.T) {
	mw := CORS(CORSConfig{AllowedOrigins: []string{"https://app.example"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Fatal("must not be called") }))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("code=%d", rec.Code)
	}
}
