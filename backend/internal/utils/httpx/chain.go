package httpx

import (
	"log/slog"
	"net/http"

	"github.com/els/backend/internal/config"
)

type ChainOptions struct {
	Logger    *slog.Logger
	Limiter   RateLimiter
	AuthPaths []string
}

func DefaultChain(h http.Handler, sec config.Security, opts ChainOptions) http.Handler {
	mws := []func(http.Handler) http.Handler{
		RequestID,
		SilenceProbes,
		AccessLog(opts.Logger),
		Recover,
		SecurityHeaders(SecurityHeadersConfig{
			HSTSMaxAge:            sec.HSTSMaxAge,
			HSTSIncludeSubdomains: sec.HSTSIncludeSub,
			HSTSPreload:           sec.HSTSPreload,
			ContentSecurityPolicy: sec.ContentSecurityPolicy,
			ReferrerPolicy:        sec.ReferrerPolicy,
		}),
		CORS(CORSConfig{
			AllowedOrigins: sec.NormalizedOrigins(),
			AllowedMethods: sec.CORSAllowedMethod,
			AllowedHeaders: sec.CORSAllowedHeader,
			ExposedHeaders: sec.CORSExposedHeader,
			MaxAge:         sec.CORSMaxAge,
		}),
		BodyLimit(BodyLimitConfig{
			Default:        sec.BodyMaxBytes,
			Upload:         sec.UploadMaxBytes,
			UploadPrefixes: sec.UploadPathPrefix,
		}),
	}
	if opts.Limiter != nil && len(opts.AuthPaths) > 0 {
		mws = append(mws, RateLimit(RateLimitConfig{
			Limiter:      opts.Limiter,
			Limit:        sec.AuthRateLimit,
			Window:       sec.AuthRateWindow,
			PathPrefixes: opts.AuthPaths,
			Methods:      []string{http.MethodPost},
			KeyPrefix:    "auth",
			Logger:       opts.Logger,
		}))
	}
	mws = append(mws, FallbackMuxErrors)
	return Chain(h, mws...)
}

func APIOptions(sec config.Security) []APIOption {
	if sec.DocsEnabled {
		return nil
	}
	return []APIOption{WithDocsDisabled()}
}
