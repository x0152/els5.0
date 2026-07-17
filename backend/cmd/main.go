package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/account"
	"github.com/els/backend/internal/application/admin"
	aiapp "github.com/els/backend/internal/application/ai"
	"github.com/els/backend/internal/application/auth"
	"github.com/els/backend/internal/application/core"
	diaryapp "github.com/els/backend/internal/application/diary"
	"github.com/els/backend/internal/application/films"
	"github.com/els/backend/internal/application/learn"
	mediaapp "github.com/els/backend/internal/application/media"
	"github.com/els/backend/internal/application/quest"
	"github.com/els/backend/internal/application/reader"
	settingsapp "github.com/els/backend/internal/application/settings"
	speechapp "github.com/els/backend/internal/application/speech"
	"github.com/els/backend/internal/application/vocab"
	"github.com/els/backend/internal/config"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/infrastructure/adapters/mailer"
	"github.com/els/backend/internal/infrastructure/adapters/redisratelimit"
	"github.com/els/backend/internal/infrastructure/adapters/s3blob"
	"github.com/els/backend/internal/infrastructure/postgres"
	cfgutil "github.com/els/backend/internal/utils/config"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/logger"
	"github.com/els/backend/internal/utils/probes"
	"github.com/els/backend/internal/utils/redisx"
)

const (
	Module  = "dev"
	Version = "0.1.0"

	minWriteTimeout = 1500 * time.Second
)

type Config struct {
	config.Global

	HTTP HTTPConfig `envPrefix:"APP_HTTP_"`
}

type HTTPConfig struct {
	Addr            string        `env:"ADDR" envDefault:":8080"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"1500s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

func (c Config) Validate() error {
	return errors.Join(
		c.Global.Validate(),
		c.HTTP.Validate(),
	)
}

func (h HTTPConfig) Validate() error {
	var errs []error
	if h.Addr == "" {
		errs = append(errs, errors.New("APP_HTTP_ADDR: must not be empty"))
	}
	if h.ReadTimeout <= 0 {
		errs = append(errs, errors.New("APP_HTTP_READ_TIMEOUT: must be > 0"))
	}
	if h.WriteTimeout <= 0 {
		errs = append(errs, errors.New("APP_HTTP_WRITE_TIMEOUT: must be > 0"))
	}
	if h.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("APP_HTTP_SHUTDOWN_TIMEOUT: must be > 0"))
	}
	return errors.Join(errs...)
}

type mounter func(api huma.API, pool *pgxpool.Pool, rdb *redis.Client, log *slog.Logger)

func main() {
	var cfg Config
	cfgutil.MustLoad(Module, &cfg)

	log := logger.New(logger.Config{
		Level:     cfg.Logger.Level,
		Format:    cfg.Logger.Format,
		Module:    Module,
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

	mux := http.NewServeMux()
	humaAPI := httpx.NewAPI(mux, Module, Version, httpx.APIOptions(cfg.Security)...)

	probes.Register(humaAPI, probes.Deps{
		Module:  Module,
		Version: Version,
		Ready: []probes.NamedCheck{
			{Name: "postgres", Check: pool.Ping},
			{Name: "redis", Check: func(ctx context.Context) error { return rdb.Ping(ctx).Err() }},
		},
	})

	storage, urls := buildMedia(cfg, log)
	if storage != nil {
		mediaapp.RegisterHTTP(mux, storage, log)
		log.Info("media proxy registered", slog.String("base", urls.Base()))
	}

	mail := mailer.New(cfg.SMTP, log)

	mounts := map[string]mounter{
		auth.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			auth.Mount(a, auth.LoadConfig(), p, r, mail)
		},
		account.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			account.Mount(a, account.LoadConfig(), p, r, l, storage, urls)
		},
		admin.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			admin.Mount(a, admin.LoadConfig(), p, r, mail)
		},
		core.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			core.Mount(ctx, a, core.LoadConfig(), p, r, l)
		},
		quest.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			quest.Mount(a, quest.LoadConfig(), p, r, l)
		},
		films.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			films.Mount(ctx, a, films.LoadConfig(), p, r, l, storage, urls)
		},
		learn.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			learn.Mount(ctx, a, learn.LoadConfig(), p, r, l, storage, urls)
		},
		vocab.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			vocab.Mount(a, mux, vocab.LoadConfig(), p, r, l, storage, urls)
		},
		reader.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			reader.Mount(ctx, a, reader.LoadConfig(), p, r, l, storage, urls)
		},
		settingsapp.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			settingsapp.Mount(a, settingsapp.LoadConfig(), p, r, l)
		},
		speechapp.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			speechapp.Mount(a, speechapp.LoadConfig(), p, r, l)
		},
		diaryapp.Name: func(a huma.API, p *pgxpool.Pool, r *redis.Client, l *slog.Logger) {
			diaryapp.Mount(a, diaryapp.LoadConfig(), p, r, l)
		},
	}

	for name, mount := range mounts {
		mount(humaAPI, pool, rdb, log.With(slog.String("module", name)))
		log.Info("module mounted", slog.String("module", name))
	}

	aiapp.Mount(humaAPI, mux, aiapp.LoadConfig(), pool, rdb, log.With(slog.String("module", aiapp.Name)), storage, urls)
	log.Info("module mounted", slog.String("module", aiapp.Name))

	limiter := redisratelimit.New(rdb, "ratelimit:auth:")
	handler := httpx.DefaultChain(mux, cfg.Security, httpx.ChainOptions{
		Logger:    log,
		Limiter:   limiter,
		AuthPaths: []string{"/api/v1/auth/"},
	})

	writeTimeout := cfg.HTTP.WriteTimeout
	if writeTimeout < minWriteTimeout {
		log.Warn(
			"increased APP_HTTP_WRITE_TIMEOUT for long-running handlers",
			slog.Duration("configured", cfg.HTTP.WriteTimeout),
			slog.Duration("effective", minWriteTimeout),
		)
		writeTimeout = minWriteTimeout
	}

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("http listening", slog.String("addr", cfg.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Info("http shutting down", slog.Duration("timeout", cfg.HTTP.ShutdownTimeout))
		shCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shCtx); err != nil {
			log.Error("shutdown failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil {
			log.Error("serve failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}
	log.Info("stopped")
}

func buildMedia(cfg Config, log *slog.Logger) (media.Storage, media.PublicURL) {
	urls := media.NewPublicURL(cfg.Media.PublicURLBase)
	store, err := s3blob.New(s3blob.Config{
		Endpoint:  cfg.S3.Endpoint,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		UseSSL:    cfg.S3.UseSSL,
		Region:    cfg.S3.Region,
	})
	if err != nil {
		log.Warn("media storage disabled: init failed", slog.String("err", err.Error()))
		return nil, urls
	}
	if err := store.EnsureBucket(context.Background(), cfg.S3.AvatarBucket); err != nil {
		log.Warn("media storage: ensure avatar bucket failed", slog.String("err", err.Error()))
	}
	return store, urls
}
