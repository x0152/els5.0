package media

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/media"
)

type handler struct {
	storage media.Storage
	logger  *slog.Logger
}

func RegisterHTTP(mux *http.ServeMux, storage media.Storage, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	h := &handler{storage: storage, logger: logger}
	mux.Handle("GET /api/v1/media/{bucket}/{key...}", h)
	mux.Handle("HEAD /api/v1/media/{bucket}/{key...}", h)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")
	if bucket == "" || key == "" {
		http.Error(w, "bucket and key are required", http.StatusBadRequest)
		return
	}

	path, err := media.NewPath(bucket + "/" + strings.TrimLeft(key, "/"))
	if err != nil {
		http.Error(w, "invalid media path", http.StatusBadRequest)
		return
	}

	reader, meta, err := h.storage.Get(r.Context(), path)
	if err != nil {
		if errors.Is(err, media.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.logger.Warn("media get failed",
			slog.String("path", path.String()),
			slog.String("err", err.Error()),
		)
		http.Error(w, "media fetch failed", http.StatusBadGateway)
		return
	}
	defer reader.Close()

	if meta.ContentType != "" {
		w.Header().Set("Content-Type", meta.ContentType)
	}
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if seeker, ok := reader.(io.ReadSeeker); ok {
		http.ServeContent(w, r, key, time.Time{}, seeker)
		return
	}

	if meta.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
	}
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	if _, err := io.Copy(w, reader); err != nil {
		h.logger.Warn("media stream failed",
			slog.String("path", path.String()),
			slog.String("err", err.Error()),
		)
	}
}
