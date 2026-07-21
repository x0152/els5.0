package listening

import (
	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	Session SessionConfig `envPrefix:"LISTENING_SESSION_"`
	LLM     LLMConfig     `envPrefix:"LLM_"`
}

type SessionConfig struct {
	KeyPrefix string `env:"KEY_PREFIX" envDefault:"session:"`
}

type LLMConfig struct {
	BaseURL string `env:"BASE_URL"`
	APIKey  string `env:"API_KEY" secret:"true"`
	Model   string `env:"MODEL"`
	Timeout int    `env:"TIMEOUT_SECONDS" envDefault:"120"`
}

func (c Config) Validate() error {
	return c.Global.Validate()
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("listening", &c)
	return c
}
