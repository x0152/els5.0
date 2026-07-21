package settings

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/settings/api"
	usecases "github.com/els/backend/internal/application/settings/use_cases"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/infrastructure/adapters/comfyui"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "settings"
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
	repo := settingsrepo.NewStore(pool)
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	defaults := cfg.Defaults()
	if err := usecases.NewSeedProvidersUseCase(repo, defaults).Execute(context.Background()); err != nil {
		logger.Warn("settings: seed ai providers", slog.String("err", err.Error()))
	}

	api.Register(humaAPI, api.Deps{
		Authenticator:      authn,
		ListProviders:      usecases.NewListProvidersUseCase(repo),
		UpdateProvider:     usecases.NewUpdateProviderUseCase(repo),
		ListModels:         usecases.NewListModelsUseCase(repo, llm.NewModelLister(), comfyui.NewModelLister(), defaults),
		GetEventProcessing: usecases.NewGetFlagUseCase(repo, domainsettings.FlagEventProcessing),
		SetEventProcessing: usecases.NewSetFlagUseCase(repo, domainsettings.FlagEventProcessing),
		GetAutoWordImages:  usecases.NewGetFlagUseCase(repo, domainsettings.FlagAutoWordImages),
		SetAutoWordImages:  usecases.NewSetFlagUseCase(repo, domainsettings.FlagAutoWordImages),
	})
}
