package workout

import (
	"time"

	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	Session SessionConfig `envPrefix:"WORKOUT_SESSION_"`
	Worker  WorkerConfig  `envPrefix:"WORKOUT_WORKER_"`
	LLM     LLMConfig     `envPrefix:"LLM_"`
}

type SessionConfig struct {
	KeyPrefix string `env:"KEY_PREFIX" envDefault:"session:"`
}

type WorkerConfig struct {
	Enabled      bool          `env:"ENABLED" envDefault:"true"`
	PlanInterval time.Duration `env:"PLAN_INTERVAL" envDefault:"1m"`
	GenInterval  time.Duration `env:"GEN_INTERVAL" envDefault:"1m"`
}

type LLMConfig struct {
	BaseURL string `env:"BASE_URL"`
	APIKey  string `env:"API_KEY" secret:"true"`
	Model   string `env:"MODEL"`
	Timeout int    `env:"TIMEOUT_SECONDS" envDefault:"600"`
}

func (c Config) Validate() error {
	return c.Global.Validate()
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("workout", &c)
	return c
}
