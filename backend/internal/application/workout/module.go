package workout

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/workout/api"
	usecases "github.com/els/backend/internal/application/workout/use_cases"
	"github.com/els/backend/internal/application/workout/worker"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	filmsrepo "github.com/els/backend/internal/infrastructure/repositories/films"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	vocabrepo "github.com/els/backend/internal/infrastructure/repositories/vocab"
	workoutrepo "github.com/els/backend/internal/infrastructure/repositories/workout"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "workout"
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

func Mount(ctx context.Context, humaAPI huma.API, cfg Config, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger) {
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	settingsStore := settingsrepo.NewStore(pool)
	resolver := providercfg.NewResolver(settingsStore, domainsettings.FeatureAnalysis,
		ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
	llmClient := llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, time.Duration(cfg.LLM.Timeout)*time.Second, resolver)

	store := workoutrepo.NewStore(pool)
	films := filmsrepo.NewStore(pool)
	generate := usecases.NewGenerateLessonUseCase(store, films, accounts, vocabrepo.NewStore(pool), store, llmClient, nil)

	api.Register(humaAPI, api.Deps{
		Authenticator:  authn,
		GetToday:       usecases.NewGetTodayUseCase(store, nil),
		GenerateLesson: generate,
		GetLesson:      usecases.NewGetLessonUseCase(store),
		SubmitStep:     usecases.NewSubmitStepUseCase(store, nil),
	})

	if cfg.Worker.Enabled {
		planFilms := usecases.NewPlanFilmsUseCase(store, films, llmClient, nil, logger)
		pregenerate := usecases.NewPregenerateUseCase(store, generate, logger)
		go worker.New(planFilms, cfg.Worker.PlanInterval, logger).Run(ctx)
		go worker.New(pregenerate, cfg.Worker.GenInterval, logger).Run(ctx)
	}
}
