package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAPI_DocsEnabledByDefault(t *testing.T) {
	mux := http.NewServeMux()
	NewAPI(mux, "test", "0.0.1")

	for _, p := range []string{"/openapi.yaml", "/openapi.json", "/docs"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code == http.StatusNotFound {
			t.Errorf("expected docs path %q to be served, got 404", p)
		}
	}
}

func TestNewAPI_WithDocsDisabled(t *testing.T) {
	mux := http.NewServeMux()
	NewAPI(mux, "test", "0.0.1", WithDocsDisabled())

	for _, p := range []string{"/openapi.yaml", "/openapi.json", "/docs"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected docs path %q to be 404, got %d", p, rec.Code)
		}
	}
}
