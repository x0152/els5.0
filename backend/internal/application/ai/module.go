package ai

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/ai/api"
	"github.com/els/backend/internal/application/ai/tools"
	usecases "github.com/els/backend/internal/application/ai/use_cases"
	learnworker "github.com/els/backend/internal/application/learn/worker"
	"github.com/els/backend/internal/domain/agent"
	"github.com/els/backend/internal/domain/media"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/imagegen"
	"github.com/els/backend/internal/infrastructure/adapters/ffmpeg"
	"github.com/els/backend/internal/infrastructure/adapters/filmvision"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	agentrepo "github.com/els/backend/internal/infrastructure/repositories/agent"
	bookrepo "github.com/els/backend/internal/infrastructure/repositories/book"
	corerepo "github.com/els/backend/internal/infrastructure/repositories/core"
	filmsrepo "github.com/els/backend/internal/infrastructure/repositories/films"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	questrepo "github.com/els/backend/internal/infrastructure/repositories/quest"
	readerrepo "github.com/els/backend/internal/infrastructure/repositories/reader"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	vocabrepo "github.com/els/backend/internal/infrastructure/repositories/vocab"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "ai"
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

func Mount(humaAPI huma.API, mux *http.ServeMux, cfg Config, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger, storage media.Storage, urls media.PublicURL) {
	repo := agentrepo.NewStore(pool)
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	provRepo := settingsrepo.NewStore(pool)
	mainResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureMain,
		ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
	client := llm.NewAgentClientWithResolver(llm.AgentConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model}, mainResolver)

	filmsStore := filmsrepo.NewStore(pool)
	var frameReader tools.FrameReader
	if storage != nil {
		visionResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureVision,
			ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
		visionClient := llm.NewVisionClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, 0, visionResolver)
		frameReader = filmvision.NewService(filmsStore, storage, ffmpeg.New(), visionClient)
	}

	ctxPlugins := []agent.ContextPlugin{
		agent.StaticContext{Prompt: agent.SystemPrompt},
		agent.IdentityContext{TZ: cfg.TZ},
		agent.ViewContext{},
	}
	imageResolver := providercfg.NewResolver(provRepo, domainsettings.FeatureImage,
		ports.AIProviderConfig{BaseURL: cfg.Image.URL, APIKey: cfg.Image.APIKey, Model: cfg.Image.Model})
	imageGen := imagegen.NewWithResolver(cfg.Image.URL, cfg.Image.APIKey, cfg.Image.Model, time.Duration(cfg.Image.Timeout)*time.Second, imageResolver)
	if ensurer, ok := storage.(media.BucketEnsurer); ok && storage != nil {
		if err := ensurer.EnsureBucket(context.Background(), cfg.Bucket); err != nil {
			logger.Warn("ai media storage: ensure bucket failed", slog.String("err", err.Error()))
		}
	}

	images := learnworker.NewImages(imageGen, storage, urls, cfg.Bucket, logger)

	toolPlugins := []agent.ToolPlugin{
		tools.NewPlugin(corerepo.NewStore(pool), vocabrepo.NewStore(pool), provRepo, images),
		tools.NewContentPlugin(tools.ContentDeps{
			Books:    bookrepo.NewStore(pool),
			Films:    filmsStore,
			Reader:   readerrepo.NewStore(pool),
			Missions: questrepo.NewStore(pool),
			Storage:  storage,
			Vision:   frameReader,
		}),
		tools.NewImagePlugin(imageGen, storage, urls, cfg.Bucket),
	}
	loop := agent.NewLoop(client, ctxPlugins, toolPlugins, nil, 0)

	service := usecases.NewService(repo, loop, client)

	deps := api.Deps{Authenticator: authn, Service: service}
	api.Register(humaAPI, deps)
	api.RegisterStream(mux, authn, service, logger)
}
