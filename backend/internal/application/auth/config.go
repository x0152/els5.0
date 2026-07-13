package auth

import (
	"errors"
	"time"

	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	HTTP     HTTPConfig     `envPrefix:"AUTH_HTTP_"`
	Session  SessionConfig  `envPrefix:"AUTH_SESSION_"`
	Password PasswordConfig `envPrefix:"AUTH_PASSWORD_"`
	Invite   config.Invite  `envPrefix:"INVITE_"`
}

type HTTPConfig struct {
	Addr            string        `env:"ADDR" envDefault:":8081"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

type SessionConfig struct {
	TTL       time.Duration `env:"TTL" envDefault:"8760h"`
	KeyPrefix string        `env:"KEY_PREFIX" envDefault:"session:"`
}

type PasswordConfig struct {
	MinLength int `env:"MIN_LENGTH" envDefault:"8"`
	MaxLength int `env:"MAX_LENGTH" envDefault:"128"`

	Argon2Memory  uint32 `env:"ARGON2_MEMORY_KIB" envDefault:"65536"`
	Argon2Time    uint32 `env:"ARGON2_TIME" envDefault:"3"`
	Argon2Threads uint8  `env:"ARGON2_THREADS" envDefault:"2"`
	Argon2SaltLen uint32 `env:"ARGON2_SALT_LEN" envDefault:"16"`
	Argon2KeyLen  uint32 `env:"ARGON2_KEY_LEN" envDefault:"32"`
}

func (c Config) Validate() error {
	return errors.Join(
		c.Global.Validate(),
		c.HTTP.Validate(),
		c.Session.Validate(),
		c.Password.Validate(c.Global.IsProd()),
		c.Invite.Validate(),
	)
}

func (h HTTPConfig) Validate() error {
	if h.Addr == "" {
		return errors.New("AUTH_HTTP_ADDR: must not be empty")
	}
	if h.ReadTimeout <= 0 {
		return errors.New("AUTH_HTTP_READ_TIMEOUT: must be > 0")
	}
	if h.WriteTimeout <= 0 {
		return errors.New("AUTH_HTTP_WRITE_TIMEOUT: must be > 0")
	}
	if h.ShutdownTimeout <= 0 {
		return errors.New("AUTH_HTTP_SHUTDOWN_TIMEOUT: must be > 0")
	}
	return nil
}

func (s SessionConfig) Validate() error {
	if s.TTL <= 0 {
		return errors.New("AUTH_SESSION_TTL: must be > 0")
	}
	return nil
}

const (
	prodMinArgon2MemoryKiB uint32 = 64 * 1024
	prodMinArgon2Time      uint32 = 2
)

func (p PasswordConfig) Validate(isProd bool) error {
	var errs []error
	if p.MinLength < 1 {
		errs = append(errs, errors.New("AUTH_PASSWORD_MIN_LENGTH: must be >= 1"))
	}
	if p.MaxLength < p.MinLength {
		errs = append(errs, errors.New("AUTH_PASSWORD_MAX_LENGTH: must be >= AUTH_PASSWORD_MIN_LENGTH"))
	}
	if p.MaxLength > 1024 {
		errs = append(errs, errors.New("AUTH_PASSWORD_MAX_LENGTH: must be <= 1024"))
	}
	if p.Argon2Memory < 8*1024 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_MEMORY_KIB: must be >= 8192"))
	}
	if p.Argon2Time < 1 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_TIME: must be >= 1"))
	}
	if p.Argon2Threads < 1 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_THREADS: must be >= 1"))
	}
	if p.Argon2SaltLen < 8 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_SALT_LEN: must be >= 8"))
	}
	if p.Argon2KeyLen < 16 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_KEY_LEN: must be >= 16"))
	}
	if isProd {
		if p.Argon2Memory < prodMinArgon2MemoryKiB {
			errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_MEMORY_KIB: must be >= 65536 in prod (OWASP)"))
		}
		if p.Argon2Time < prodMinArgon2Time {
			errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_TIME: must be >= 2 in prod"))
		}
	}
	return errors.Join(errs...)
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("auth", &c)
	return c
}
