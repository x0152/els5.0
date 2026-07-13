package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/els/backend/internal/domain/shared"
)

func withCapturedLogger(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return buf, func() { slog.SetDefault(prev) }
}

type logRecord struct {
	Level   string `json:"level"`
	Msg     string `json:"msg"`
	Err     string `json:"err"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	ErrType string `json:"err_type"`
	Stack   string `json:"stack"`
}

func parseLast(t *testing.T, buf *bytes.Buffer) logRecord {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("no log records captured: %q", buf.String())
	}
	var rec logRecord
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &rec); err != nil {
		t.Fatalf("failed to parse log record: %v; line=%q", err, lines[len(lines)-1])
	}
	return rec
}

func TestLogResponseError_SkipsSuccess(t *testing.T) {
	buf, restore := withCapturedLogger(t)
	defer restore()

	LogResponseError(context.Background(), nil, http.StatusOK, "")
	LogResponseError(context.Background(), errors.New("boom"), http.StatusOK, "")

	if buf.Len() != 0 {
		t.Fatalf("expected no log records for status<400, got %q", buf.String())
	}
}

func TestLogResponseError_ClientError_WarnNoStack(t *testing.T) {
	buf, restore := withCapturedLogger(t)
	defer restore()

	err := fmt.Errorf("user %q not found: %w", "bob", shared.ErrNotFound)
	LogResponseError(context.Background(), err, http.StatusNotFound, CodeNotFound)

	rec := parseLast(t, buf)
	if rec.Level != "WARN" {
		t.Fatalf("expected WARN for 4xx, got %q", rec.Level)
	}
	if rec.Msg != "request error" {
		t.Fatalf("expected msg 'request error', got %q", rec.Msg)
	}
	if !strings.Contains(rec.Err, "user \"bob\" not found") {
		t.Fatalf("expected wrapped err message, got %q", rec.Err)
	}
	if rec.Status != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Status)
	}
	if rec.Code != string(CodeNotFound) {
		t.Fatalf("expected code NOT_FOUND, got %q", rec.Code)
	}
	if rec.Stack != "" {
		t.Fatalf("did not expect a stack for 4xx, got %q", rec.Stack)
	}
}

func TestLogResponseError_ServerError_ErrorWithStack(t *testing.T) {
	buf, restore := withCapturedLogger(t)
	defer restore()

	err := errors.New("database is on fire")
	LogResponseError(context.Background(), err, http.StatusInternalServerError, CodeInternal)

	rec := parseLast(t, buf)
	if rec.Level != "ERROR" {
		t.Fatalf("expected ERROR for 5xx, got %q", rec.Level)
	}
	if rec.Err != "database is on fire" {
		t.Fatalf("expected err text, got %q", rec.Err)
	}
	if rec.Stack == "" {
		t.Fatalf("expected stack trace for 5xx, got empty")
	}
	if !strings.Contains(rec.Stack, "errorlog_test.go") && !strings.Contains(rec.Stack, "errorlog.go") {
		t.Fatalf("stack looks unexpected: %s", rec.Stack)
	}
}

func TestLogResponseError_SkipsOpaqueInternalEnvelope(t *testing.T) {
	buf, restore := withCapturedLogger(t)
	defer restore()

	opaque := &Error{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "internal server error",
	}
	LogResponseError(context.Background(), opaque, http.StatusInternalServerError, CodeInternal)

	if buf.Len() != 0 {
		t.Fatalf("expected no log for opaque 500 envelope, got %q", buf.String())
	}
}

func TestLogResponseError_StackFromHttpxError(t *testing.T) {
	buf, restore := withCapturedLogger(t)
	defer restore()

	err := Wrap(errors.New("root cause"), http.StatusServiceUnavailable, CodeUnavailable, "downstream is down")
	if len(err.Stack) == 0 {
		t.Fatalf("expected Wrap to capture stack for 5xx, got none")
	}

	LogResponseError(context.Background(), err, err.Status, err.Code)

	rec := parseLast(t, buf)
	if rec.Level != "ERROR" {
		t.Fatalf("expected ERROR for 5xx, got %q", rec.Level)
	}
	if rec.Stack == "" {
		t.Fatalf("expected stack, got empty")
	}
	if rec.ErrType != "httpx.Error" {
		t.Fatalf("expected err_type 'httpx.Error', got %q", rec.ErrType)
	}
}

func TestNewError_CapturesStackOnly5xx(t *testing.T) {
	ok4xx := NewError(http.StatusBadRequest, CodeBadRequest, "nope")
	if len(ok4xx.Stack) != 0 {
		t.Fatalf("4xx NewError should not capture stack, got %d bytes", len(ok4xx.Stack))
	}

	err5xx := NewError(http.StatusInternalServerError, CodeInternal, "boom")
	if len(err5xx.Stack) == 0 {
		t.Fatalf("5xx NewError must capture stack")
	}
	if !strings.Contains(string(err5xx.Stack), "TestNewError_CapturesStackOnly5xx") {
		t.Fatalf("stack does not mention the calling test; got:\n%s", err5xx.Stack)
	}
}
