package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/els/backend/internal/application/core"
	"github.com/els/backend/internal/infrastructure/postgres"
	cfgutil "github.com/els/backend/internal/utils/config"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/logger"
	"github.com/els/backend/internal/utils/redisx"
)

func main() {
	cfg := core.LoadConfig()

	log := logger.New(logger.Config{
		Level:     cfg.Logger.Level,
		Format:    cfg.Logger.Format,
		Module:    core.Name,
		AddSource: cfg.Logger.AddSource,
	})
	slog.SetDefault(log)

	cfgutil.Dump(log, cfg)
	httpx.InstallHumaErrorHandler()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, postgres.Config{
		DSN:             cfg.Postgres.DSN().Raw(),
		MaxConns:        int32(cfg.Postgres.MaxOpenConns),
		MinConns:        int32(cfg.Postgres.MaxIdleConns),
		ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
		ConnectTimeout:  cfg.Postgres.ConnectTimeout,
	})
	if err != nil {
		log.Error("postgres connect failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	rdb, err := redisx.New(ctx, redisx.Config{
		Addr:           cfg.Redis.Addr,
		Password:       cfg.Redis.Password,
		DB:             cfg.Redis.DB,
		ConnectTimeout: cfg.Redis.ConnectTimeout,
	})
	if err != nil {
		log.Error("redis connect failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer rdb.Close()

	app := core.New(ctx, cfg, pool, rdb, log)
	if err := app.Run(ctx); err != nil {
		log.Error("run failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("stopped")
}
