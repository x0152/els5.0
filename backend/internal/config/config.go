package config

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

type Global struct {
	App           App           `envPrefix:"APP_"`
	Logger        Logger        `envPrefix:"LOG_"`
	Postgres      Postgres      `envPrefix:"POSTGRES_"`
	Redis         Redis         `envPrefix:"REDIS_"`
	S3            S3            `envPrefix:"S3_"`
	Media         Media         `envPrefix:"MEDIA_"`
	SMTP          SMTP          `envPrefix:"SMTP_"`
	Security      Security      `envPrefix:"SECURITY_"`
	Impersonation Impersonation `envPrefix:"IMPERSONATION_"`
}

type Impersonation struct {
	Enabled bool `env:"ENABLED" envDefault:"false"`
}

func (i Impersonation) Validate(_ string) error {
	return nil
}

type App struct {
	Env  string `env:"ENV" envDefault:"dev"`
	Name string `env:"NAME" envDefault:"els-data-platform-backend"`
}

type Logger struct {
	Level     string `env:"LEVEL" envDefault:"info"`
	Format    string `env:"FORMAT" envDefault:"json"`
	AddSource bool   `env:"ADD_SOURCE" envDefault:"true"`
}

type Redis struct {
	Addr           string        `env:"ADDR" envDefault:"localhost:6379"`
	Password       string        `env:"PASSWORD" secret:"true"`
	DB             int           `env:"DB" envDefault:"0"`
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT" envDefault:"5s"`
}

type S3 struct {
	Endpoint     string `env:"ENDPOINT" envDefault:"localhost:9000"`
	AccessKey    string `env:"ACCESS_KEY,required" secret:"true"`
	SecretKey    string `env:"SECRET_KEY,required" secret:"true"`
	UseSSL       bool   `env:"USE_SSL" envDefault:"false"`
	Region       string `env:"REGION" envDefault:"us-east-1"`
	AvatarBucket string `env:"AVATAR_BUCKET" envDefault:"avatars"`
}

func (s S3) Validate() error {
	var errs []error
	if strings.TrimSpace(s.Endpoint) == "" {
		errs = append(errs, errors.New("S3_ENDPOINT: must not be empty"))
	}
	if strings.TrimSpace(s.AccessKey) == "" {
		errs = append(errs, errors.New("S3_ACCESS_KEY: must not be empty"))
	}
	if strings.TrimSpace(s.SecretKey) == "" {
		errs = append(errs, errors.New("S3_SECRET_KEY: must not be empty"))
	}
	if strings.TrimSpace(s.Region) == "" {
		errs = append(errs, errors.New("S3_REGION: must not be empty"))
	}
	if strings.TrimSpace(s.AvatarBucket) == "" {
		errs = append(errs, errors.New("S3_AVATAR_BUCKET: must not be empty"))
	}
	return errors.Join(errs...)
}

type Media struct {
	PublicURLBase string `env:"PUBLIC_URL_BASE" envDefault:"/api/v1/media"`
}

func (m Media) Validate() error {
	if strings.TrimSpace(m.PublicURLBase) == "" {
		return errors.New("MEDIA_PUBLIC_URL_BASE: must not be empty")
	}
	if !strings.HasPrefix(m.PublicURLBase, "/") && !strings.Contains(m.PublicURLBase, "://") {
		return fmt.Errorf("MEDIA_PUBLIC_URL_BASE: must be absolute path or full URL, got %q", m.PublicURLBase)
	}
	return nil
}

type Postgres struct {
	Host            string        `env:"HOST,required"`
	Port            int           `env:"PORT" envDefault:"5432"`
	User            string        `env:"USER,required"`
	Password        string        `env:"PASSWORD,required" secret:"true"`
	Database        string        `env:"DATABASE,required"`
	SSLMode         string        `env:"SSLMODE" envDefault:"disable"`
	Timezone        string        `env:"TIMEZONE" envDefault:"Europe/Moscow"`
	MaxOpenConns    int32         `env:"MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns    int32         `env:"MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"30m"`
	ConnectTimeout  time.Duration `env:"CONNECT_TIMEOUT" envDefault:"5s"`
}

type PostgresDSN string

func (d PostgresDSN) Raw() string { return string(d) }

func (d PostgresDSN) String() string {
	u, err := url.Parse(string(d))
	if err != nil {
		return "<invalid-dsn>"
	}
	if u.User != nil {
		if _, ok := u.User.Password(); ok {
			u.User = url.UserPassword(u.User.Username(), "***")
		}
	}
	return u.String()
}

func (d PostgresDSN) LogValue() any { return d.String() }

func (p Postgres) DSN() PostgresDSN {
	v := url.Values{}
	v.Set("sslmode", p.SSLMode)
	v.Set("timezone", p.Timezone)
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     fmt.Sprintf("%s:%d", p.Host, p.Port),
		Path:     "/" + p.Database,
		RawQuery: v.Encode(),
	}
	return PostgresDSN(u.String())
}

var (
	validEnvs       = []string{"dev", "stage", "prod"}
	validLogLevels  = []string{"debug", "info", "warn", "error"}
	validLogFormats = []string{"json", "text"}
)

func (g Global) Validate() error {
	return errors.Join(
		g.App.Validate(),
		g.Logger.Validate(),
		g.Postgres.Validate(),
		g.Redis.Validate(),
		g.S3.Validate(),
		g.Media.Validate(),
		g.SMTP.Validate(),
		g.Security.Validate(),
		g.Impersonation.Validate(g.App.Env),
	)
}

func (g Global) IsProd() bool { return g.App.Env == "prod" }

func (r Redis) Validate() error {
	var errs []error
	if r.Addr == "" {
		errs = append(errs, errors.New("REDIS_ADDR: must not be empty"))
	}
	if r.DB < 0 {
		errs = append(errs, fmt.Errorf("REDIS_DB: must be >= 0, got %d", r.DB))
	}
	if r.ConnectTimeout <= 0 {
		errs = append(errs, errors.New("REDIS_CONNECT_TIMEOUT: must be > 0"))
	}
	return errors.Join(errs...)
}

func (a App) Validate() error {
	if !slices.Contains(validEnvs, a.Env) {
		return fmt.Errorf("APP_ENV: must be one of %v, got %q", validEnvs, a.Env)
	}
	if a.Name == "" {
		return errors.New("APP_NAME: must not be empty")
	}
	return nil
}

func (l Logger) Validate() error {
	var errs []error
	if !slices.Contains(validLogLevels, strings.ToLower(l.Level)) {
		errs = append(errs, fmt.Errorf("LOG_LEVEL: must be one of %v, got %q", validLogLevels, l.Level))
	}
	if !slices.Contains(validLogFormats, strings.ToLower(l.Format)) {
		errs = append(errs, fmt.Errorf("LOG_FORMAT: must be one of %v, got %q", validLogFormats, l.Format))
	}
	return errors.Join(errs...)
}

func (p Postgres) Validate() error {
	var errs []error
	if p.Port <= 0 || p.Port > 65535 {
		errs = append(errs, fmt.Errorf("POSTGRES_PORT: must be 1-65535, got %d", p.Port))
	}
	if p.Timezone == "" {
		errs = append(errs, errors.New("POSTGRES_TIMEZONE: must not be empty"))
	}
	if p.MaxOpenConns <= 0 {
		errs = append(errs, errors.New("POSTGRES_MAX_OPEN_CONNS: must be > 0"))
	}
	if p.MaxIdleConns < 0 {
		errs = append(errs, errors.New("POSTGRES_MAX_IDLE_CONNS: must be >= 0"))
	}
	if p.MaxIdleConns > p.MaxOpenConns {
		errs = append(errs, errors.New("POSTGRES_MAX_IDLE_CONNS must be <= POSTGRES_MAX_OPEN_CONNS"))
	}
	if p.ConnMaxLifetime <= 0 {
		errs = append(errs, errors.New("POSTGRES_CONN_MAX_LIFETIME: must be > 0"))
	}
	if p.ConnectTimeout <= 0 {
		errs = append(errs, errors.New("POSTGRES_CONNECT_TIMEOUT: must be > 0"))
	}
	return errors.Join(errs...)
}
