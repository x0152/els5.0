package core

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/core/api"
	usecases "github.com/els/backend/internal/application/core/use_cases"
	"github.com/els/backend/internal/application/core/worker"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	"github.com/els/backend/internal/infrastructure/adapters/settingsflag"
	corerepo "github.com/els/backend/internal/infrastructure/repositories/core"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/openapi"
	"github.com/els/backend/internal/utils/probes"
	"github.com/els/backend/internal/utils/timex"
)

const (
	Name    = "core"
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

	store := corerepo.NewStore(pool)
	settingsStore := settingsrepo.NewStore(pool)
	mainResolver := providercfg.NewResolver(settingsStore, domainsettings.FeatureMain,
		ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
	llmClient := llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, time.Duration(cfg.LLM.Timeout)*time.Second, mainResolver)
	gate := settingsflag.NewEventProcessingGate(settingsStore)
	if cfg.Worker.Enabled {
		processEvents := usecases.NewProcessEventsUseCase(store, llmClient, gate, logger, cfg.Worker.Batch)
		retryFailed := usecases.NewRetryFailedEventsUseCase(store, llmClient, gate, logger, cfg.Worker.Batch)
		enrichCatalog := usecases.NewEnrichCatalogUseCase(store, llmClient, gate, logger, cfg.Worker.Batch)
		go worker.New(processEvents, cfg.Worker.Interval, logger).Run(ctx)
		go worker.New(retryFailed, cfg.Worker.Interval, logger).Run(ctx)
		go worker.New(enrichCatalog, cfg.Worker.EnrichInterval, logger).Run(ctx)
	} else {
		logger.Info("core event processing disabled", slog.String("module", Name))
	}

	api.Register(humaAPI, api.Deps{
		Authenticator:    authn,
		IngestEvents:     usecases.NewIngestEventsUseCase(store, timex.System()),
		MarkUnclear:      usecases.NewMarkUnclearUseCase(store, timex.System()),
		ListEvents:       usecases.NewListEventsUseCase(store),
		ListCatalog:      usecases.NewListCatalogUseCase(store),
		ListDictionaries: usecases.NewListDictionariesUseCase(),
		WipeData:         usecases.NewWipeDataUseCase(store),
		DeleteRows:       usecases.NewDeleteRowsUseCase(store),
	})
}

type App struct {
	cfg    Config
	logger *slog.Logger
	srv    *http.Server
}

func New(ctx context.Context, cfg Config, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger) *App {
	mux := http.NewServeMux()
	humaAPI := httpx.NewAPI(mux, Name, Version, httpx.APIOptions(cfg.Security)...)

	probes.Register(humaAPI, probes.Deps{
		Module:  Name,
		Version: Version,
		Ready: []probes.NamedCheck{
			{Name: "postgres", Check: pool.Ping},
			{Name: "redis", Check: func(ctx context.Context) error { return rdb.Ping(ctx).Err() }},
		},
	})

	Mount(ctx, humaAPI, cfg, pool, rdb, logger)

	handler := httpx.DefaultChain(mux, cfg.Security, httpx.ChainOptions{Logger: logger})

	return &App{
		cfg:    cfg,
		logger: logger,
		srv: &http.Server{
			Addr:         cfg.HTTP.Addr,
			Handler:      handler,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
	}
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("http listening", slog.String("addr", a.cfg.HTTP.Addr))
		if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("http shutting down",
			slog.Duration("timeout", a.cfg.HTTP.ShutdownTimeout),
		)
		shCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()
		return a.srv.Shutdown(shCtx)
	case err := <-errCh:
		return err
	}
}
