package httpx

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/els/backend/internal/utils/reqctx"
)

const HeaderRequestID = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(HeaderRequestID)
		if rid == "" {
			rid = generateRequestID()
		}
		w.Header().Set(HeaderRequestID, rid)
		ctx := reqctx.WithRequestID(r.Context(), rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := newObservedWriter(w)
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			slog.ErrorContext(r.Context(), "panic recovered",
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			)
			if ow.HeaderWritten() || ow.hijacked {
				return
			}
			WriteError(ow, r, &Error{
				Status:  http.StatusInternalServerError,
				Code:    CodeInternal,
				Message: "internal server error",
			})
		}()
		next.ServeHTTP(ow, r)
	})
}

func FallbackMuxErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fw := &fallbackWriter{ResponseWriter: w, req: r}
		next.ServeHTTP(fw, r)
	})
}

type fallbackWriter struct {
	http.ResponseWriter
	req           *http.Request
	headerWritten bool
	rewrite       bool
}

func (f *fallbackWriter) WriteHeader(status int) {
	if f.headerWritten {
		return
	}
	f.headerWritten = true

	ct := f.ResponseWriter.Header().Get("Content-Type")
	isMuxFallback := (status == http.StatusNotFound || status == http.StatusMethodNotAllowed) &&
		(ct == "" || strings.HasPrefix(ct, "text/plain"))
	if !isMuxFallback {
		f.ResponseWriter.WriteHeader(status)
		return
	}

	f.rewrite = true
	for k := range f.ResponseWriter.Header() {
		if strings.EqualFold(k, "Content-Type") || strings.EqualFold(k, "X-Content-Type-Options") {
			f.ResponseWriter.Header().Del(k)
		}
	}

	var herr *Error
	switch status {
	case http.StatusNotFound:
		herr = NewError(http.StatusNotFound, CodeNotFound, "route not found")
	default:
		herr = NewError(http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
	}
	WriteError(f.ResponseWriter, f.req, herr)
}

func (f *fallbackWriter) Write(p []byte) (int, error) {
	if !f.headerWritten {
		f.WriteHeader(http.StatusOK)
	}
	if f.rewrite {
		return len(p), nil
	}
	return f.ResponseWriter.Write(p)
}

func (f *fallbackWriter) Flush() {
	if fl, ok := f.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (f *fallbackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := f.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	f.headerWritten = true
	return h.Hijack()
}

func (f *fallbackWriter) Unwrap() http.ResponseWriter { return f.ResponseWriter }

func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func generateRequestID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err == nil {
		return "req_" + hex.EncodeToString(b[:])
	}
	return "req_fb_" + strconv.FormatInt(time.Now().UnixNano(), 16)
}
