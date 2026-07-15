package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Security struct {
	BodyMaxBytes      int64         `env:"BODY_MAX_BYTES" envDefault:"1048576"`
	UploadMaxBytes    int64         `env:"UPLOAD_MAX_BYTES" envDefault:"5368709120"`
	UploadPathPrefix  []string      `env:"UPLOAD_PATH_PREFIX" envSeparator:"," envDefault:"/api/v1/account,/api/v1/films,/api/v1/reader"`
	CORSAllowedOrigin []string      `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:""`
	CORSAllowedMethod []string      `env:"CORS_ALLOWED_METHODS" envSeparator:"," envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	CORSAllowedHeader []string      `env:"CORS_ALLOWED_HEADERS" envSeparator:"," envDefault:"Authorization,Content-Type,X-Request-ID,Idempotency-Key"`
	CORSExposedHeader []string      `env:"CORS_EXPOSED_HEADERS" envSeparator:"," envDefault:"X-Request-ID"`
	CORSMaxAge        time.Duration `env:"CORS_MAX_AGE" envDefault:"600s"`

	HSTSMaxAge            time.Duration `env:"HSTS_MAX_AGE" envDefault:"31536000s"`
	HSTSIncludeSub        bool          `env:"HSTS_INCLUDE_SUBDOMAINS" envDefault:"true"`
	HSTSPreload           bool          `env:"HSTS_PRELOAD" envDefault:"false"`
	ContentSecurityPolicy string        `env:"CONTENT_SECURITY_POLICY" envDefault:"default-src 'none'; frame-ancestors 'none'"`
	ReferrerPolicy        string        `env:"REFERRER_POLICY" envDefault:"no-referrer"`

	DocsEnabled bool `env:"DOCS_ENABLED" envDefault:"true"`

	AuthRateLimit       int           `env:"AUTH_RATE_LIMIT" envDefault:"10"`
	AuthRateWindow      time.Duration `env:"AUTH_RATE_WINDOW" envDefault:"60s"`
	AuthLockoutAttempts int           `env:"AUTH_LOCKOUT_ATTEMPTS" envDefault:"10"`
	AuthLockoutWindow   time.Duration `env:"AUTH_LOCKOUT_WINDOW" envDefault:"15m"`
}

func (s Security) Validate() error {
	var errs []error
	if s.BodyMaxBytes <= 0 {
		errs = append(errs, errors.New("SECURITY_BODY_MAX_BYTES: must be > 0"))
	}
	if s.UploadMaxBytes <= 0 {
		errs = append(errs, errors.New("SECURITY_UPLOAD_MAX_BYTES: must be > 0"))
	}
	if s.UploadMaxBytes < s.BodyMaxBytes {
		errs = append(errs, errors.New("SECURITY_UPLOAD_MAX_BYTES: must be >= SECURITY_BODY_MAX_BYTES"))
	}
	for _, o := range s.CORSAllowedOrigin {
		if v := strings.TrimSpace(o); v != "" && v != "*" && !strings.Contains(v, "://") {
			errs = append(errs, fmt.Errorf("SECURITY_CORS_ALLOWED_ORIGINS: %q must be a full URL with scheme or '*'", v))
		}
	}
	if s.CORSMaxAge < 0 {
		errs = append(errs, errors.New("SECURITY_CORS_MAX_AGE: must be >= 0"))
	}
	if s.HSTSMaxAge < 0 {
		errs = append(errs, errors.New("SECURITY_HSTS_MAX_AGE: must be >= 0"))
	}
	if s.AuthRateLimit <= 0 {
		errs = append(errs, errors.New("SECURITY_AUTH_RATE_LIMIT: must be > 0"))
	}
	if s.AuthRateWindow <= 0 {
		errs = append(errs, errors.New("SECURITY_AUTH_RATE_WINDOW: must be > 0"))
	}
	if s.AuthLockoutAttempts <= 0 {
		errs = append(errs, errors.New("SECURITY_AUTH_LOCKOUT_ATTEMPTS: must be > 0"))
	}
	if s.AuthLockoutWindow <= 0 {
		errs = append(errs, errors.New("SECURITY_AUTH_LOCKOUT_WINDOW: must be > 0"))
	}
	return errors.Join(errs...)
}

func (s Security) NormalizedOrigins() []string {
	out := make([]string, 0, len(s.CORSAllowedOrigin))
	for _, o := range s.CORSAllowedOrigin {
		if v := strings.TrimSpace(o); v != "" {
			out = append(out, v)
		}
	}
	return out
}
