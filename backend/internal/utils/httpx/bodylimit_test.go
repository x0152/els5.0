package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimit_RejectsOverDefault(t *testing.T) {
	var readErr error
	mw := BodyLimit(BodyLimitConfig{Default: 10})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(strings.Repeat("a", 100)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !IsBodyTooLarge(readErr) {
		t.Fatalf("expected MaxBytesError, got %v", readErr)
	}
}

func TestBodyLimit_AllowsUnderDefault(t *testing.T) {
	var n int
	mw := BodyLimit(BodyLimitConfig{Default: 100})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		n = len(body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if n != 5 {
		t.Fatalf("read=%d", n)
	}
}

func TestBodyLimit_UploadPathUsesLargerLimit(t *testing.T) {
	var readErr error
	mw := BodyLimit(BodyLimitConfig{Default: 5, Upload: 1000, UploadPrefixes: []string{"/api/v1/account"}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/picture", strings.NewReader(strings.Repeat("a", 500)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if readErr != nil {
		t.Fatalf("expected no error on upload path, got %v", readErr)
	}
}

func TestBodyLimit_GET_NoLimit(t *testing.T) {
	var readErr error
	mw := BodyLimit(BodyLimitConfig{Default: 5})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", strings.NewReader("xxxxxxxxxxxxxxxx"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if readErr != nil {
		t.Fatalf("GET should bypass body limit, got %v", readErr)
	}
}
