package bindings

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"

	authusecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/application/grid_engine/api"
	"github.com/els/backend/internal/application/grid_engine/gridspec"
	"github.com/els/backend/internal/application/grid_engine/lookups"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/database"
)

func Administrators(
	admins admin.Repository,
	_ iam.AccountRepository,
	invites *authusecases.InviteAccountUseCase,
) gridspec.Binding {
	cfg := gridspec.Config[*admin.Administrator]{
		BasePath:  "/api/v1/administrators",
		Tag:       "administrator",
		Summary:   "Administrators",
		Authorize: iam.RequireGlobalAdmin,
		Grid:      admin.NewGrid,
		CRUD: gridspec.CRUD[*admin.Administrator]{
			List: func(ctx context.Context, actor *iam.Actor, page gridspec.Page) ([]*admin.Administrator, int64, error) {
				return admins.List(ctx, admin.VisibilityFor(actor), page.Limit, page.Offset)
			},
			GetByID: func(ctx context.Context, _ *iam.Actor, id string) (*admin.Administrator, error) {
				parsed, err := vo.ParseID(id)
				if err != nil {
					return nil, shared.Validation(fmt.Errorf("administrator.id: %w", err))
				}
				return admins.GetByID(ctx, admin.ID{ID: parsed})
			},
			Create: func(ctx context.Context, _ *iam.Actor, data map[grid.ColumnID]any) (*admin.Administrator, error) {
				email, firstName, lastName, err := requireAccountFields(data)
				if err != nil {
					return nil, err
				}
				acc, err := invites.Execute(ctx, authusecases.InviteAccountCommand{
					Email:     email,
					FirstName: firstName,
					LastName:  lastName,
				})
				if err != nil {
					return nil, err
				}
				a, err := admin.NewAdministratorNow(admin.NewID(), acc)
				if err != nil {
					return nil, err
				}
				if err := admins.Create(ctx, a); err != nil {
					return nil, err
				}
				return a, nil
			},
			Update: func(ctx context.Context, a *admin.Administrator) error { return admins.Update(ctx, a) },
			AfterUpdate: inviteOnNoAuthToPending(invites, func(a *admin.Administrator) iam.AccountSide {
				return a.AccountSide
			}),
			Version: func(a *admin.Administrator) int64 { return a.Version() },
			Delete: func(ctx context.Context, _ *iam.Actor, id string) error {
				parsed, err := vo.ParseID(id)
				if err != nil {
					return shared.Validation(fmt.Errorf("administrator.id: %w", err))
				}
				return admins.Delete(ctx, admin.ID{ID: parsed})
			},
		},
	}

	return gridspec.Binding{
		BasePath: cfg.BasePath,
		Tag:      cfg.Tag,
		Summary:  cfg.Summary,
		Register: func(humaAPI huma.API, auth *authx.Authenticator, resolver *lookups.Resolver, tx database.TxRunner) {
			api.RegisterGrid(humaAPI, auth, resolver, tx, cfg)
		},
	}
}

func requireString(data map[grid.ColumnID]any, id grid.ColumnID) (string, error) {
	v, ok := data[id]
	if !ok || v == nil {
		return "", shared.Validation(fmt.Errorf("column %q: required", id))
	}
	s, ok := v.(string)
	if !ok {
		return "", shared.Validation(fmt.Errorf("column %q: must be a string", id))
	}
	return s, nil
}

func requireAccountFields(data map[grid.ColumnID]any) (email, firstName, lastName string, err error) {
	email, err = requireString(data, iam.ColAccountEmail)
	if err != nil {
		return "", "", "", err
	}
	firstName, err = requireString(data, iam.ColAccountFirstName)
	if err != nil {
		return "", "", "", err
	}
	lastName, err = requireString(data, iam.ColAccountLastName)
	if err != nil {
		return "", "", "", err
	}
	return email, firstName, lastName, nil
}
