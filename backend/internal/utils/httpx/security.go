package httpx

import (
	"fmt"
	"net/http"
	"time"
)

type SecurityHeadersConfig struct {
	HSTSMaxAge            time.Duration
	HSTSIncludeSubdomains bool
	HSTSPreload           bool
	ContentSecurityPolicy string
	ReferrerPolicy        string
}

func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	hsts := buildHSTS(cfg)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			if cfg.ReferrerPolicy != "" {
				h.Set("Referrer-Policy", cfg.ReferrerPolicy)
			}
			if cfg.ContentSecurityPolicy != "" {
				h.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}
			if hsts != "" && (r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
				h.Set("Strict-Transport-Security", hsts)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func buildHSTS(cfg SecurityHeadersConfig) string {
	if cfg.HSTSMaxAge <= 0 {
		return ""
	}
	out := fmt.Sprintf("max-age=%d", int(cfg.HSTSMaxAge.Seconds()))
	if cfg.HSTSIncludeSubdomains {
		out += "; includeSubDomains"
	}
	if cfg.HSTSPreload {
		out += "; preload"
	}
	return out
}
