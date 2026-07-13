package httpx

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareChain_PreservesFlusherAndHijacker(t *testing.T) {
	var sawFlusher, sawHijacker bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); ok {
			sawFlusher = true
		}
		if _, ok := w.(http.Hijacker); ok {
			sawHijacker = true
		}
		w.WriteHeader(http.StatusOK)
	})

	chain := Chain(handler,
		RequestID,
		AccessLog(slog.New(slog.NewTextHandler(io.Discard, nil))),
		Recover,
		FallbackMuxErrors,
	)

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := &hijackableRecorder{ResponseRecorder: httptest.NewRecorder()}
	chain.ServeHTTP(rec, req)

	if !sawFlusher {
		t.Fatal("http.Flusher not accessible through middleware chain")
	}
	if !sawHijacker {
		t.Fatal("http.Hijacker not accessible through middleware chain")
	}
}

func TestMiddlewareChain_ResponseControllerFlush(t *testing.T) {
	var flushed bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		if err := rc.Flush(); err == nil {
			flushed = true
		}
	})

	chain := Chain(handler,
		RequestID,
		AccessLog(slog.New(slog.NewTextHandler(io.Discard, nil))),
		Recover,
		FallbackMuxErrors,
	)

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	chain.ServeHTTP(httptest.NewRecorder(), req)

	if !flushed {
		t.Fatal("ResponseController.Flush() could not reach underlying writer")
	}
}

var errHijackUnsupportedForTest = errors.New("hijack not supported in recorder")

type hijackableRecorder struct {
	*httptest.ResponseRecorder
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errHijackUnsupportedForTest
}

func (h *hijackableRecorder) Flush() {
	h.ResponseRecorder.Flush()
}

func TestFallbackMuxErrors_Rewrites404(t *testing.T) {
	mux := http.NewServeMux()
	chain := FallbackMuxErrors(mux)

	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("want json content-type, got %q", ct)
	}
	var body struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.OK {
		t.Fatal("expected ok=false")
	}
	if body.Error.Code != string(CodeNotFound) {
		t.Fatalf("expected code NOT_FOUND, got %q", body.Error.Code)
	}
}

func TestFallbackMuxErrors_Rewrites405(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /only", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	chain := FallbackMuxErrors(mux)

	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("want json content-type, got %q", ct)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Error.Code != string(CodeMethodNotAllowed) {
		t.Fatalf("expected METHOD_NOT_ALLOWED, got %q", body.Error.Code)
	}
}

func TestFallbackMuxErrors_PassesThroughJSON404(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"custom":"not-found"}`))
	})
	chain := FallbackMuxErrors(handler)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"custom":"not-found"}` {
		t.Fatalf("body was rewritten: %q", body)
	}
}

func TestFallbackMuxErrors_PassesThroughSuccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})
	chain := FallbackMuxErrors(handler)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("body corrupted: %q", rec.Body.String())
	}
}

func TestFallbackMuxErrors_StreamsBytesWithoutBuffering(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("chunk-1\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		_, _ = w.Write([]byte("chunk-2\n"))
	})
	chain := FallbackMuxErrors(handler)

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if rec.Body.String() != "chunk-1\nchunk-2\n" {
		t.Fatalf("body corrupted: %q", rec.Body.String())
	}
}
