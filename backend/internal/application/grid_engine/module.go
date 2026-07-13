package gridengine

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	authusecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/application/grid_engine/api"
	"github.com/els/backend/internal/application/grid_engine/bindings"
	"github.com/els/backend/internal/application/grid_engine/gridspec"
	"github.com/els/backend/internal/application/grid_engine/lookups"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/infrastructure/postgres"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator  *authx.Authenticator
	Pool           *pgxpool.Pool
	Accounts       iam.AccountRepository
	Invites        *authusecases.InviteAccountUseCase
	Administrators admin.Repository
}

func Mount(humaAPI huma.API, deps Deps) {
	sources := []lookups.Source{
		lookups.NewIAMAccountsSource(deps.Accounts),
	}
	resolver := lookups.NewResolver(sources...)

	var binds []gridspec.Binding
	if deps.Administrators != nil {
		binds = append(binds, bindings.Administrators(deps.Administrators, deps.Accounts, deps.Invites))
	}

	api.Register(humaAPI, api.Deps{
		Authenticator: deps.Authenticator,
		Lookups:       resolver,
		TxRunner:      postgres.NewTxRunner(deps.Pool),
		Bindings:      binds,
	})
}
