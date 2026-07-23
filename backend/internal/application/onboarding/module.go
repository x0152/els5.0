package onboarding

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/onboarding/api"
	usecases "github.com/els/backend/internal/application/onboarding/use_cases"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	onboardingrepo "github.com/els/backend/internal/infrastructure/repositories/onboarding"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/openapi"
)

const (
	Name    = "onboarding"
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

	repo := onboardingrepo.NewStore(pool)

	api.Register(humaAPI, api.Deps{
		Authenticator: authn,
		GetProgress:   usecases.NewGetProgressUseCase(repo, nil),
		AckItems:      usecases.NewAckItemsUseCase(repo, nil),
		GetTours:      usecases.NewGetToursUseCase(repo),
		MarkTour:      usecases.NewMarkTourUseCase(repo, nil),
		ResetTours:    usecases.NewResetToursUseCase(repo),
	})
}
