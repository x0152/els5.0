package quest

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/quest/api"
	"github.com/els/backend/internal/application/quest/runtime"
	usecases "github.com/els/backend/internal/application/quest/use_cases"
	"github.com/els/backend/internal/domain/media"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/bothub"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	"github.com/els/backend/internal/infrastructure/adapters/s3blob"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	questrepo "github.com/els/backend/internal/infrastructure/repositories/quest"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "quest"
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
) {
	store := questrepo.NewStore(pool)
	if err := store.FailStaleGenerating(context.Background()); err != nil {
		logger.Warn("quest: fail stale missions", slog.String("err", err.Error()))
	}
	profiles := questrepo.NewProfileStore(pool)
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	provRepo := settingsrepo.NewStore(pool)
	mainFallback := ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model}
	mainResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureMain, mainFallback)
	analysisResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureAnalysis, mainFallback)

	llmBase := time.Duration(cfg.LLM.Timeout) * time.Second
	llmClient := llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, llmBase*8, mainResolver)
	grammarClient := llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, llmBase*8, analysisResolver)
	gateway := runtime.NewLLMGateway(llmClient, grammarClient, llmBase, logger)

	imageResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureImage,
		ports.AIProviderConfig{BaseURL: cfg.Image.URL, APIKey: cfg.Image.APIKey, Model: cfg.Image.Model})
	imageGen := bothub.NewWithResolver(cfg.Image.URL, cfg.Image.APIKey, cfg.Image.Model, time.Duration(cfg.Image.Timeout)*time.Second, imageResolver)

	storage, urls := buildStorage(cfg, logger)
	images := runtime.NewImages(imageGen, storage, cfg.Bucket, gateway, store, logger)
	generator := runtime.NewGenerator(gateway, store, profiles, images, accounts, urls, logger)
	dialog := runtime.NewDialog(gateway, store, profiles, accounts, images, logger)

	api.Register(humaAPI, api.Deps{
		Authenticator:    authn,
		CreateMission:    usecases.NewCreateMissionUseCase(store, generator),
		ListMissions:     usecases.NewListMissionsUseCase(store),
		GetMission:       usecases.NewGetMissionUseCase(store, dialog, logger),
		StartRespond:     usecases.NewStartRespondUseCase(dialog),
		SuggestNative:    usecases.NewSuggestNativeReplyUseCase(dialog),
		ResetMission:     usecases.NewResetMissionUseCase(store),
		RegenerateImages: usecases.NewRegenerateImagesUseCase(store, images),
		DeleteMission:    usecases.NewDeleteMissionUseCase(store, images),
		MediaURLs:        urls,
		MediaBucket:      cfg.Bucket,
	})
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
		logger.Warn("quest media storage disabled: init failed", slog.String("err", err.Error()))
		return nil, urls
	}
	if err := store.EnsureBucket(context.Background(), cfg.Bucket); err != nil {
		logger.Warn("quest media storage: ensure bucket failed", slog.String("err", err.Error()))
	}
	return store, urls
}
