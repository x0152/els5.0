package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status, body := MapError(r.Context(), err)
	LogResponseError(r.Context(), err, status, body.Code)
	writeJSON(r, w, status, ErrorResponse{
		OK:   false,
		Err:  &body,
		Meta: metaFromCtx(r.Context(), nil),
	})
}

func writeJSON(r *http.Request, w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.WarnContext(r.Context(), "write json response failed",
			slog.String("err", err.Error()),
		)
	}
}
