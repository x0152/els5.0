package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	usecases "github.com/els/backend/internal/application/ai/use_cases"
	"github.com/els/backend/internal/domain/agent"
	authx "github.com/els/backend/internal/utils/auth"
)

func toView(in *struct {
	App    string            `json:"app"`
	Screen string            `json:"screen"`
	Title  string            `json:"title"`
	Info   string            `json:"info"`
	IDs    map[string]string `json:"ids"`
	State  map[string]any    `json:"state"`
}) *agent.View {
	if in == nil {
		return nil
	}
	state := make(map[string]string, len(in.State))
	for k, v := range in.State {
		state[k] = fmt.Sprint(v)
	}
	return &agent.View{
		App:    in.App,
		Screen: in.Screen,
		Title:  in.Title,
		Info:   in.Info,
		IDs:    in.IDs,
		State:  state,
	}
}

type streamHandler struct {
	authn   *authx.Authenticator
	service *usecases.Service
	logger  *slog.Logger
}

func RegisterStream(mux *http.ServeMux, authn *authx.Authenticator, service *usecases.Service, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	mux.Handle("POST /api/v1/ai/stream", &streamHandler{authn: authn, service: service, logger: logger})
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actor, _, err := h.authn.Authenticate(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Message    string `json:"message"`
		Regenerate bool   `json:"regenerate"`
		View       *struct {
			App    string            `json:"app"`
			Screen string            `json:"screen"`
			Title  string            `json:"title"`
			Info   string            `json:"info"`
			IDs    map[string]string `json:"ids"`
			State  map[string]any    `json:"state"`
		} `json:"view"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (!body.Regenerate && strings.TrimSpace(body.Message) == "") {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}
	view := toView(body.View)
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

	emit := func(ev agent.Event) {
		switch ev.Type {
		case agent.EventText:
			send("text", map[string]any{"delta": ev.TextDelta})
		case agent.EventToolStart:
			send("tool_start", map[string]any{"id": ev.Step.ID, "tool": ev.Step.Tool, "label": ev.Step.Label, "icon": ev.Step.Icon, "args": ev.Step.Args})
		case agent.EventToolEnd:
			send("tool_end", map[string]any{"id": ev.StepID, "tool": ev.Step.Tool, "result": ev.ToolResult})
		case agent.EventError:
			send("error", map[string]any{"message": ev.ErrMessage})
		case agent.EventDone:
			send("done", map[string]any{"finish": string(ev.FinishReason), "model": ev.Model, "total_tokens": ev.Usage.TotalTokens})
		}
	}

	var streamErr error
	if body.Regenerate {
		streamErr = h.service.Regenerate(r.Context(), actor, view, emit)
	} else {
		streamErr = h.service.Stream(r.Context(), actor, body.Message, view, emit)
	}
	if streamErr != nil {
		h.logger.Warn("ai stream failed", slog.String("err", streamErr.Error()))
		send("error", map[string]any{"message": streamErr.Error()})
	}
}
