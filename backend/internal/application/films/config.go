package films

import (
	"github.com/els/backend/internal/config"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	Session  SessionConfig `envPrefix:"FILMS_SESSION_"`
	Bucket   string        `env:"FILMS_S3_BUCKET" envDefault:"films"`
	TempDir  string        `env:"FILMS_TEMP_DIR" envDefault:"/tmp"`
	SpacyURL string        `env:"SPACY_URL" envDefault:"http://localhost:8000"`
}

type SessionConfig struct {
	KeyPrefix string `env:"KEY_PREFIX" envDefault:"session:"`
}

func (c Config) Validate() error {
	return c.Global.Validate()
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("films", &c)
	return c
}
