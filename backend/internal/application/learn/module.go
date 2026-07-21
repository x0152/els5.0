package learn

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/learn/api"
	usecases "github.com/els/backend/internal/application/learn/use_cases"
	"github.com/els/backend/internal/application/learn/worker"
	"github.com/els/backend/internal/domain/media"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/imagegen"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	bookrepo "github.com/els/backend/internal/infrastructure/repositories/book"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	practicerepo "github.com/els/backend/internal/infrastructure/repositories/practice"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "learn"
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

func Mount(ctx context.Context, humaAPI huma.API, cfg Config, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger, storage media.Storage, urls media.PublicURL) {
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	chapters := bookrepo.NewStore(pool)
	if err := chapters.FailStaleGenerating(ctx); err != nil {
		logger.Warn("learn: fail stale chapters", slog.String("err", err.Error()))
	}
	books, err := seedBooks()
	if err != nil {
		logger.Error("learn seed books load failed", slog.String("err", err.Error()))
	}
	if seed, err := seedChapters(); err != nil {
		logger.Error("learn seed load failed", slog.String("err", err.Error()))
	} else if err := usecases.NewSeedChaptersUseCase(chapters, books, seed).Execute(ctx); err != nil {
		logger.Error("learn seed failed", slog.String("err", err.Error()))
	}

	provRepo := settingsrepo.NewStore(pool)
	imageResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureImage,
		ports.AIProviderConfig{BaseURL: cfg.Image.URL, APIKey: cfg.Image.APIKey, Model: cfg.Image.Model})
	imageProvider := imagegen.NewWithResolver(cfg.Image.URL, cfg.Image.APIKey, cfg.Image.Model, time.Duration(cfg.Image.Timeout)*time.Second, imageResolver)
	if ensurer, ok := storage.(media.BucketEnsurer); ok {
		if err := ensurer.EnsureBucket(ctx, cfg.Bucket); err != nil {
			logger.Warn("learn: ensure illustration bucket failed", slog.String("err", err.Error()))
		}
	}
	images := worker.NewImages(imageProvider, storage, urls, cfg.Bucket, logger)

	variants := practicerepo.NewVariantStore(pool)
	if err := variants.FailStaleGenerating(ctx); err != nil {
		logger.Warn("learn: fail stale variants", slog.String("err", err.Error()))
	}
	progress := practicerepo.NewProgressStore(pool)
	sources := worker.NewSources(pool)
	mainResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureMain,
		ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
	service := worker.NewService(llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, time.Duration(cfg.LLM.Timeout)*time.Second, mainResolver))
	variantWorker := worker.NewVariants(variants, sources, service, logger)
	chapterWorker := worker.NewChapters(chapters, service, logger)

	api.Register(humaAPI, api.Deps{
		Authenticator: authn,

		ListBooks:       usecases.NewListBooksUseCase(chapters),
		ListChapters:    usecases.NewListChaptersUseCase(chapters),
		GetChapter:      usecases.NewGetChapterUseCase(chapters),
		CreateChapter:   usecases.NewCreateChapterUseCase(chapters),
		UpdateChapter:   usecases.NewUpdateChapterUseCase(chapters),
		DeleteChapter:   usecases.NewDeleteChapterUseCase(chapters),
		GenerateChapter: usecases.NewGenerateChapterUseCase(chapters, chapterWorker),

		EnsureIllustration: usecases.NewEnsureIllustrationUseCase(images),

		ListVariants:    usecases.NewListVariantsUseCase(variants),
		GenerateVariant: usecases.NewGenerateVariantUseCase(variants, variantWorker),
		DeleteVariant:   usecases.NewDeleteVariantUseCase(variants, progress),
		GetProgress:     usecases.NewGetProgressUseCase(progress),
		SaveProgress:    usecases.NewSaveProgressUseCase(progress),
		ResetProgress:   usecases.NewResetProgressUseCase(progress),
		CheckFree:       usecases.NewCheckFreeUseCase(sources, service),
	})
}
