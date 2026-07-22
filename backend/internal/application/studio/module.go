package studio

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/studio/api"
	usecases "github.com/els/backend/internal/application/studio/use_cases"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	studiorepo "github.com/els/backend/internal/infrastructure/repositories/studio"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/openapi"
	"github.com/els/backend/internal/utils/probes"
)

const (
	Name    = "studio"
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

func Mount(humaAPI huma.API, cfg Config, pool *pgxpool.Pool, rdb *redis.Client, _ *slog.Logger) {
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	settingsStore := settingsrepo.NewStore(pool)
	resolver := providercfg.NewResolver(settingsStore, domainsettings.FeatureAnalysis,
		ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
	llmClient := llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, time.Duration(cfg.LLM.Timeout)*time.Second, resolver)

	repo := studiorepo.NewStore(pool)

	api.Register(humaAPI, api.Deps{
		Authenticator: authn,
		ListAreas:     usecases.NewListAreasUseCase(repo, nil),
		CreateArea:    usecases.NewCreateAreaUseCase(repo, nil),
		DeleteArea:    usecases.NewDeleteAreaUseCase(repo),
		ListItems:     usecases.NewListItemsUseCase(repo),
		AddItem:       usecases.NewAddItemUseCase(repo, llmClient, nil),
		CaptureItem:   usecases.NewCaptureItemUseCase(repo, llmClient, nil),
		DeleteItem:    usecases.NewDeleteItemUseCase(repo),
		MarkSkill:     usecases.NewMarkSkillUseCase(repo, nil),
		PassReview:    usecases.NewPassReviewUseCase(repo, nil),
		RegenExample:  usecases.NewRegenExampleUseCase(repo, llmClient),
		RegenTask:     usecases.NewRegenTaskUseCase(repo, llmClient),
		CheckReply:    usecases.NewCheckReplyUseCase(repo, llmClient, nil),
	})
}

type App struct {
	cfg    Config
	logger *slog.Logger
	srv    *http.Server
}

func New(cfg Config, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger) *App {
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

	Mount(humaAPI, cfg, pool, rdb, logger)

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
