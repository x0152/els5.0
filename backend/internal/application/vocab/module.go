package vocab

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	learnworker "github.com/els/backend/internal/application/learn/worker"
	"github.com/els/backend/internal/application/vocab/api"
	usecases "github.com/els/backend/internal/application/vocab/use_cases"
	"github.com/els/backend/internal/application/vocab/worker"
	"github.com/els/backend/internal/domain/media"
	domainsettings "github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/bothub"
	"github.com/els/backend/internal/infrastructure/adapters/llm"
	"github.com/els/backend/internal/infrastructure/adapters/providercfg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	lexiconrepo "github.com/els/backend/internal/infrastructure/repositories/lexicon"
	settingsrepo "github.com/els/backend/internal/infrastructure/repositories/settings"
	vocabrepo "github.com/els/backend/internal/infrastructure/repositories/vocab"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "vocab"
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
	store := vocabrepo.NewStore(pool)
	lex := lexiconrepo.NewStore(pool)
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	settingsStore := settingsrepo.NewStore(pool)
	analysisResolver := providercfg.NewResolver(settingsStore, domainsettings.FeatureAnalysis,
		ports.AIProviderConfig{BaseURL: cfg.LLM.BaseURL, APIKey: cfg.LLM.APIKey, Model: cfg.LLM.Model})
	llmClient := llm.NewWithResolver(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, time.Duration(cfg.LLM.Timeout)*time.Second, analysisResolver)

	imageResolver := providercfg.NewResolver(settingsStore, domainsettings.FeatureImage,
		ports.AIProviderConfig{BaseURL: cfg.Image.URL, APIKey: cfg.Image.APIKey, Model: cfg.Image.Model})
	imageProvider := bothub.NewWithResolver(cfg.Image.URL, cfg.Image.APIKey, cfg.Image.Model, time.Duration(cfg.Image.Timeout)*time.Second, imageResolver)
	images := learnworker.NewImages(imageProvider, storage, urls, cfg.Bucket, logger)

	practiceSessions := vocabrepo.NewPracticeSessionStore(rdb)
	practiceWorker := worker.NewPractice(practiceSessions, llmClient, logger)

	analyze := usecases.NewAnalyzeUseCase(llmClient, lex, store)
	api.Register(humaAPI, api.Deps{
		Authenticator:    authn,
		AddUnit:          usecases.NewAddUnitUseCase(store, llmClient, settingsStore, images),
		Analyze:          analyze,
		Occurrences:      usecases.NewOccurrencesUseCase(lex),
		ListUnits:        usecases.NewListUnitsUseCase(store),
		UpdateStatus:     usecases.NewUpdateStatusUseCase(store),
		DeleteUnit:       usecases.NewDeleteUnitUseCase(store),
		GeneratePractice: usecases.NewGeneratePracticeUseCase(store, practiceSessions, practiceWorker),
		GetPractice:      usecases.NewGetPracticeUseCase(practiceSessions),
		SaveProgress:     usecases.NewSavePracticeProgressUseCase(practiceSessions),
		CheckPractice:    usecases.NewCheckPracticeUseCase(llmClient),
		GenerateCards:    usecases.NewGenerateCardsUseCase(store, storage, urls, cfg.Bucket),
		AnswerCard:       usecases.NewAnswerCardUseCase(store, nil),
		DueCards:         usecases.NewDueCardsUseCase(store, nil),
	})
	api.RegisterStream(mux, authn, analyze, logger)
}
