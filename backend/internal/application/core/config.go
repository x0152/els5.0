package core

import (
	"errors"
	"time"

	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	HTTP    HTTPConfig    `envPrefix:"CORE_HTTP_"`
	Session SessionConfig `envPrefix:"CORE_SESSION_"`
	Worker  WorkerConfig  `envPrefix:"CORE_WORKER_"`
	LLM     LLMConfig     `envPrefix:"LLM_"`
}

type HTTPConfig struct {
	Addr            string        `env:"ADDR" envDefault:":8090"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

type SessionConfig struct {
	KeyPrefix string `env:"KEY_PREFIX" envDefault:"session:"`
}

type WorkerConfig struct {
	Enabled        bool          `env:"ENABLED" envDefault:"true"`
	Interval       time.Duration `env:"INTERVAL" envDefault:"5s"`
	EnrichInterval time.Duration `env:"ENRICH_INTERVAL" envDefault:"20s"`
	Batch          int           `env:"BATCH" envDefault:"5"`
}

type LLMConfig struct {
	BaseURL string `env:"BASE_URL"`
	APIKey  string `env:"API_KEY" secret:"true"`
	Model   string `env:"MODEL"`
	Timeout int    `env:"TIMEOUT_SECONDS" envDefault:"600"`
}

func (c Config) Validate() error {
	return errors.Join(
		c.Global.Validate(),
		c.HTTP.Validate(),
	)
}

func (h HTTPConfig) Validate() error {
	if h.Addr == "" {
		return errors.New("CORE_HTTP_ADDR: must not be empty")
	}
	if h.ReadTimeout <= 0 {
		return errors.New("CORE_HTTP_READ_TIMEOUT: must be > 0")
	}
	if h.WriteTimeout <= 0 {
		return errors.New("CORE_HTTP_WRITE_TIMEOUT: must be > 0")
	}
	if h.ShutdownTimeout <= 0 {
		return errors.New("CORE_HTTP_SHUTDOWN_TIMEOUT: must be > 0")
	}
	return nil
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("core", &c)
	return c
}
