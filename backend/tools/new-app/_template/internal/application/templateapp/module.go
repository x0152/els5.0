package templateapp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/templateapp/api"
	usecases "github.com/els/backend/internal/application/templateapp/use_cases"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/openapi"
	"github.com/els/backend/internal/utils/probes"
)

const (
	Name    = "templateapp"
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

	api.Register(humaAPI, api.Deps{
		Authenticator: authn,
		Echo:          usecases.NewEchoUseCase(),
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
