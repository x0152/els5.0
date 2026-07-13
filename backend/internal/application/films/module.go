package films

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/films/api"
	usecases "github.com/els/backend/internal/application/films/use_cases"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/infrastructure/adapters/ffmpeg"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	"github.com/els/backend/internal/infrastructure/adapters/spacy"
	filmsrepo "github.com/els/backend/internal/infrastructure/repositories/films"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	lexiconrepo "github.com/els/backend/internal/infrastructure/repositories/lexicon"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "films"
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
	store := filmsrepo.NewStore(pool)
	lex := lexiconrepo.NewStore(pool)
	analyzer := spacy.NewClient(cfg.SpacyURL)
	accounts := iamrepo.NewAccountRepo(pool)
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	authn := authx.New(sessions, accounts)

	deps := api.Deps{
		Authenticator: authn,
		ListFilms:     usecases.NewListFilmsUseCase(store),
		GetFilm:       usecases.NewGetFilmUseCase(store),
		DeleteFilm:    usecases.NewDeleteFilmUseCase(store, storage, lex, logger),
		SaveProgress:  usecases.NewSaveProgressUseCase(store, nil),
		ListSeries:    usecases.NewListSeriesUseCase(store),
		MediaURLs:     urls,
		TempDir:       cfg.TempDir,
	}
	if storage != nil {
		if ensurer, ok := storage.(media.BucketEnsurer); ok {
			if err := ensurer.EnsureBucket(ctx, cfg.Bucket); err != nil {
				logger.Warn("films media storage: ensure bucket failed", slog.String("err", err.Error()))
			}
		}
		deps.UploadFilm = usecases.NewUploadFilmUseCase(store, storage, ffmpeg.New(), analyzer, lex, cfg.Bucket, cfg.TempDir, logger)
		deps.UpdateFilm = usecases.NewUpdateFilmUseCase(store, storage, cfg.Bucket)
		deps.UpdateSeries = usecases.NewUpdateSeriesUseCase(store, storage, cfg.Bucket)
	}

	api.Register(humaAPI, deps)
}
