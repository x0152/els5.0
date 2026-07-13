package reader

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/reader/api"
	usecases "github.com/els/backend/internal/application/reader/use_cases"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/infrastructure/adapters/pandoc"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	"github.com/els/backend/internal/infrastructure/adapters/spacy"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	lexiconrepo "github.com/els/backend/internal/infrastructure/repositories/lexicon"
	readerrepo "github.com/els/backend/internal/infrastructure/repositories/reader"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "reader"
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
	store := readerrepo.NewStore(pool)
	lex := lexiconrepo.NewStore(pool)
	analyzer := spacy.NewClient(cfg.SpacyURL)
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	deps := api.Deps{
		Authenticator:   authn,
		ListBooks:       usecases.NewListBooksUseCase(store),
		GetBook:         usecases.NewGetBookUseCase(store),
		SaveProgress:    usecases.NewSaveProgressUseCase(store),
		DeleteBook:      usecases.NewDeleteBookUseCase(store, storage, lex, logger),
		ListCollections: usecases.NewListCollectionsUseCase(store),
		MediaURLs:       urls,
		TempDir:         cfg.TempDir,
	}
	if storage != nil {
		if ensurer, ok := storage.(media.BucketEnsurer); ok {
			if err := ensurer.EnsureBucket(ctx, cfg.Bucket); err != nil {
				logger.Warn("reader media storage: ensure bucket failed", slog.String("err", err.Error()))
			}
		}
		deps.UploadBook = usecases.NewUploadBookUseCase(store, storage, pandoc.New(), urls, analyzer, lex, cfg.Bucket, cfg.TempDir, logger)
		deps.ImportArticle = usecases.NewImportArticleUseCase(deps.UploadBook, cfg.TempDir)
		deps.UpdateBook = usecases.NewUpdateBookUseCase(store, storage, cfg.Bucket)
		deps.UpdateCollection = usecases.NewUpdateCollectionUseCase(store, storage, cfg.Bucket)
	}

	api.Register(humaAPI, deps)
}
