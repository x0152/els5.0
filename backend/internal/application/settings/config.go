package settings

import (
	"github.com/els/backend/internal/config"
	domain "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	cfgutil "github.com/els/backend/internal/utils/config"
)

type Config struct {
	config.Global

	Session  SessionConfig `envPrefix:"AI_SESSION_"`
	Main     ProviderEnv   `envPrefix:"LLM_"`
	Analysis ProviderEnv   `envPrefix:"LLM_ANALYSIS_"`
	Vision   ProviderEnv   `envPrefix:"LLM_VISION_"`
	Image    ImageEnv      `envPrefix:"IMAGE_"`
}

type SessionConfig struct {
	KeyPrefix string `env:"KEY_PREFIX" envDefault:"session:"`
}

type ProviderEnv struct {
	BaseURL string `env:"BASE_URL"`
	APIKey  string `env:"API_KEY" secret:"true"`
	Model   string `env:"MODEL"`
}

type ImageEnv struct {
	URL    string `env:"API_URL"`
	APIKey string `env:"API_KEY" secret:"true"`
	Model  string `env:"MODEL"`
}

// Defaults returns the env-provided seed for each AI provider. Analysis and
// vision fall back to the main provider when not configured separately.
func (c Config) Defaults() map[domain.Feature]ports.AIProviderConfig {
	main := ports.AIProviderConfig{BaseURL: c.Main.BaseURL, APIKey: c.Main.APIKey, Model: c.Main.Model}
	analysis := ports.AIProviderConfig{BaseURL: c.Analysis.BaseURL, APIKey: c.Analysis.APIKey, Model: c.Analysis.Model}
	vision := ports.AIProviderConfig{BaseURL: c.Vision.BaseURL, APIKey: c.Vision.APIKey, Model: c.Vision.Model}
	if analysis.IsEmpty() {
		analysis = main
	}
	if vision.IsEmpty() {
		vision = main
	}
	return map[domain.Feature]ports.AIProviderConfig{
		domain.FeatureMain:     main,
		domain.FeatureAnalysis: analysis,
		domain.FeatureVision:   vision,
		domain.FeatureImage:    {BaseURL: c.Image.URL, APIKey: c.Image.APIKey, Model: c.Image.Model},
	}
}

func (c Config) Validate() error {
	return c.Global.Validate()
}

func LoadConfig() Config {
	var c Config
	cfgutil.MustLoad("settings", &c)
	return c
}
