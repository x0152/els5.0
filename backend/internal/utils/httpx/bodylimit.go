package httpx

import (
	"errors"
	"net/http"
	"strings"
)

type BodyLimitConfig struct {
	Default        int64
	Upload         int64
	UploadPrefixes []string
}

func BodyLimit(cfg BodyLimitConfig) func(http.Handler) http.Handler {
	if cfg.Default <= 0 && cfg.Upload <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			limit := cfg.Default
			if cfg.Upload > 0 {
				for _, prefix := range cfg.UploadPrefixes {
					if prefix != "" && strings.HasPrefix(r.URL.Path, prefix) {
						limit = cfg.Upload
						break
					}
				}
			}
			if limit <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func IsBodyTooLarge(err error) bool {
	if err == nil {
		return false
	}
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}
