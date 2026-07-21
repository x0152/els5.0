package speech

import (
	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	Session    SessionConfig `envPrefix:"SPEECH_SESSION_"`
	LLM        LLMConfig     `envPrefix:"LLM_"`
	ServiceURL string        `env:"SPEECH_URL" envDefault:"http://localhost:8001"`
	TTSURL     string        `env:"TTS_URL" envDefault:"http://localhost:8002"`
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
	cfgutil.MustLoad("speech", &c)
	return c
}
