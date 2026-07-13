package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	usecases "github.com/els/backend/internal/application/vocab/use_cases"
	authx "github.com/els/backend/internal/utils/auth"
)

type analyzeStreamHandler struct {
	authn   *authx.Authenticator
	analyze *usecases.AnalyzeUseCase
	logger  *slog.Logger
}

func RegisterStream(mux *http.ServeMux, authn *authx.Authenticator, analyze *usecases.AnalyzeUseCase, logger *slog.Logger) {
	if analyze == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	mux.Handle("POST /api/v1/vocab/analyze/stream", &analyzeStreamHandler{authn: authn, analyze: analyze, logger: logger})
}

func (h *analyzeStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actor, _, err := h.authn.Authenticate(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Text    string `json:"text"`
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Text) == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	send := func(event string, data any) {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		flusher.Flush()
	}

	emit := func(it usecases.AnalyzeItem) { send("item", toAnalyzeItemOutput(it)) }
	if err := h.analyze.Stream(r.Context(), actor, body.Text, body.Context, emit); err != nil {
		h.logger.Warn("vocab analyze stream failed", slog.String("err", err.Error()))
		send("error", map[string]any{"message": err.Error()})
		return
	}
	send("done", map[string]any{})
}
