package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSecurityHeaders_AlwaysSetsBaseHeaders(t *testing.T) {
	mw := SecurityHeaders(SecurityHeadersConfig{
		ContentSecurityPolicy: "default-src 'none'",
		ReferrerPolicy:        "no-referrer",
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options=%q", got)
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options=%q", got)
	}
	if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Errorf("Referrer-Policy=%q", got)
	}
	if got := rec.Header().Get("Content-Security-Policy"); got != "default-src 'none'" {
		t.Errorf("CSP=%q", got)
	}
}

func TestSecurityHeaders_HSTSOnlyOverHTTPS(t *testing.T) {
	mw := SecurityHeaders(SecurityHeadersConfig{HSTSMaxAge: time.Hour, HSTSIncludeSubdomains: true})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	t.Run("plain http: no HSTS", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Header().Get("Strict-Transport-Security") != "" {
			t.Fatal("HSTS must not be set on plain http")
		}
	})

	t.Run("forwarded https: HSTS present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		got := rec.Header().Get("Strict-Transport-Security")
		if !strings.Contains(got, "max-age=3600") || !strings.Contains(got, "includeSubDomains") {
			t.Fatalf("unexpected HSTS=%q", got)
		}
	})
}
