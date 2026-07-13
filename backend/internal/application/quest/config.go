package quest

import (
	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	Session SessionConfig `envPrefix:"QUEST_SESSION_"`
	LLM     LLMConfig     `envPrefix:"LLM_"`
	Image   ImageConfig
	Bucket  string `env:"QUEST_S3_BUCKET" envDefault:"quest"`
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

type ImageConfig struct {
	URL     string `env:"IMAGE_API_URL"`
	APIKey  string `env:"IMAGE_API_KEY" secret:"true"`
	Model   string `env:"IMAGE_MODEL"`
	Timeout int    `env:"IMAGE_TIMEOUT_SECONDS" envDefault:"180"`
}

func (c Config) Validate() error {
	return c.Global.Validate()
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("quest", &c)
	return c
}
