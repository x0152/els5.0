package templateapp

import (
	"errors"
	"time"

	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	HTTP    HTTPConfig    `envPrefix:"TEMPLATEAPP_HTTP_"`
	Session SessionConfig `envPrefix:"TEMPLATEAPP_SESSION_"`
}

type HTTPConfig struct {
	Addr            string        `env:"ADDR" envDefault:":8099"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

type SessionConfig struct {
	KeyPrefix string `env:"KEY_PREFIX" envDefault:"session:"`
}

func (c Config) Validate() error {
	return errors.Join(
		c.Global.Validate(),
		c.HTTP.Validate(),
	)
}

func (h HTTPConfig) Validate() error {
	if h.Addr == "" {
		return errors.New("TEMPLATEAPP_HTTP_ADDR: must not be empty")
	}
	if h.ReadTimeout <= 0 {
		return errors.New("TEMPLATEAPP_HTTP_READ_TIMEOUT: must be > 0")
	}
	if h.WriteTimeout <= 0 {
		return errors.New("TEMPLATEAPP_HTTP_WRITE_TIMEOUT: must be > 0")
	}
	if h.ShutdownTimeout <= 0 {
		return errors.New("TEMPLATEAPP_HTTP_SHUTDOWN_TIMEOUT: must be > 0")
	}
	return nil
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("templateapp", &c)
	return c
}
