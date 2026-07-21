package writing

import (
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/writing/api"
	usecases "github.com/els/backend/internal/application/writing/use_cases"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "writing"
	Version = "0.1.0"
)

func init() {
	openapi.Register(openapi.Module{
		Name:    Name,
		Version: Version,
		Register: func(a huma.API) {
			api.Register(a, api.Deps{})
		},
	})
}

func Mount(humaAPI huma.API, cfg Config, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger) {
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	settingsStore := settingsrepo.NewStore(pool)
	resolver := providercfg.NewResolver(settingsStore, domainsettings.FeatureAnalysis,
		ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
	llmClient := llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, time.Duration(cfg.LLM.Timeout)*time.Second, resolver)

	api.Register(humaAPI, api.Deps{
		Authenticator:     authn,
		TrainerCheck:      usecases.NewTrainerCheckUseCase(llmClient),
		GenerateSituation: usecases.NewGenerateSituationUseCase(llmClient),
	})
}
