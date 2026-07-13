package httpx

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposedHeaders []string
	MaxAge         time.Duration
}

func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	if len(cfg.AllowedOrigins) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	allowAny := false
	originSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			allowAny = true
			continue
		}
		originSet[o] = struct{}{}
	}
	methods := strings.Join(defaultIfEmpty(cfg.AllowedMethods, []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions}), ", ")
	headers := strings.Join(defaultIfEmpty(cfg.AllowedHeaders, []string{"Authorization", "Content-Type"}), ", ")
	exposed := strings.Join(cfg.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(int(cfg.MaxAge.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			allowed := allowAny
			if !allowed {
				_, allowed = originSet[origin]
			}
			if !allowed {
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			h := w.Header()
			if allowAny {
				h.Set("Access-Control-Allow-Origin", "*")
			} else {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
			}
			if exposed != "" {
				h.Set("Access-Control-Expose-Headers", exposed)
			}
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				h.Set("Access-Control-Allow-Methods", methods)
				h.Set("Access-Control-Allow-Headers", headers)
				h.Set("Access-Control-Max-Age", maxAge)
				h.Add("Vary", "Access-Control-Request-Method")
				h.Add("Vary", "Access-Control-Request-Headers")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func defaultIfEmpty(in, def []string) []string {
	if len(in) == 0 {
		return def
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if v := strings.TrimSpace(s); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}
