package account

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/account/api"
	usecases "github.com/els/backend/internal/application/account/use_cases"
	mediaapp "github.com/els/backend/internal/application/media"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/infrastructure/adapters/contentsniff"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	"github.com/els/backend/internal/infrastructure/adapters/s3blob"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/openapi"
	"github.com/els/backend/internal/utils/probes"
)

const (
	Name    = "account"
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

func Mount(
	humaAPI huma.API,
	cfg Config,
	pool *pgxpool.Pool,
	rdb *redis.Client,
	logger *slog.Logger,
	storage media.Storage,
	urls media.PublicURL,
) {
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)

	authn := authx.New(sessions, accounts)
	meUC := usecases.NewMeUseCase()
	updateProfileUC := usecases.NewUpdateProfileUseCase(accounts)
	listAppsUC := usecases.NewListAppsUseCase(nil)

	var uploadPictureUC *usecases.UploadAccountPictureUseCase
	if storage != nil {
		uploadPictureUC = usecases.NewUploadAccountPictureUseCase(
			accounts,
			storage,
			contentsniff.New(),
			urls,
			usecases.UploadAccountPictureConfig{
				Bucket:       cfg.S3.AvatarBucket,
				MaxSizeBytes: cfg.Picture.MaxSizeMB * 1024 * 1024,
			},
			logger,
		)
	}

	api.Register(humaAPI, api.Deps{
		Authenticator:        authn,
		Me:                   meUC,
		UpdateProfile:        updateProfileUC,
		ListApps:             listAppsUC,
		UploadAccountPicture: uploadPictureUC,
		ImpersonationEnabled: cfg.Impersonation.Enabled,
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

	storage, urls := buildStorage(cfg, logger)
	if storage != nil {
		mediaapp.RegisterHTTP(mux, storage, logger)
	}

	Mount(humaAPI, cfg, pool, rdb, logger, storage, urls)

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

func buildStorage(cfg Config, logger *slog.Logger) (media.Storage, media.PublicURL) {
	urls := media.NewPublicURL(cfg.Media.PublicURLBase)
	store, err := s3blob.New(s3blob.Config{
		Endpoint:  cfg.S3.Endpoint,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		UseSSL:    cfg.S3.UseSSL,
		Region:    cfg.S3.Region,
	})
	if err != nil {
		if logger != nil {
			logger.Warn("s3blob disabled: init failed", slog.String("err", err.Error()))
		}
		return nil, urls
	}
	if err := store.EnsureBucket(context.Background(), cfg.S3.AvatarBucket); err != nil && logger != nil {
		logger.Warn("s3blob: ensure avatar bucket failed", slog.String("err", err.Error()))
	}
	return store, urls
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
