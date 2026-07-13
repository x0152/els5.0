package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/els/backend/internal/application/grid_engine/gridspec"
	"github.com/els/backend/internal/application/grid_engine/lookups"
	usecases "github.com/els/backend/internal/application/grid_engine/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/database"
)

type Deps struct {
	Authenticator *authx.Authenticator
	Lookups       *lookups.Resolver
	TxRunner      database.TxRunner
	Bindings      []gridspec.Binding
}

func Register(api huma.API, deps Deps) {
	for _, b := range deps.Bindings {
		if b.Register == nil {
			continue
		}
		b.Register(api, deps.Authenticator, deps.Lookups, deps.TxRunner)
	}
}

func RegisterGrid[E any](api huma.API, auth *authx.Authenticator, resolver *lookups.Resolver, tx database.TxRunner, cfg gridspec.Config[E]) {
	if cfg.Grid == nil {
		panic("grid_engine/api: Config.Grid is required")
	}
	if cfg.CRUD.List == nil {
		panic("grid_engine/api: Config.CRUD.List is required")
	}
	if cfg.BasePath == "" {
		panic("grid_engine/api: Config.BasePath is required")
	}
	if cfg.Tag == "" {
		panic("grid_engine/api: Config.Tag is required")
	}

	describeUC := usecases.NewDescribeGridUseCase(cfg, resolver)
	applyUC := usecases.NewApplyGridUseCase(cfg, tx)
	lookupUC := usecases.NewLookupGridUseCase(cfg, resolver)

	tagTitle := title(cfg.Tag)
	summary := cfg.Summary
	if summary == "" {
		summary = tagTitle + " grid"
	}
	summaryLower := strings.ToLower(summary)

	authx.Authed(api, auth, huma.Operation{
		OperationID: "describe" + tagTitle + "Grid",
		Method:      http.MethodGet,
		Path:        cfg.BasePath + "/grid",
		Summary:     "Describe " + summaryLower,
		Tags:        []string{cfg.Tag},
	}, func(ctx context.Context, actor *iam.Actor, in *DescribeGridInput) (DescribeGridOutput, error) {
		res, err := describeUC.Execute(ctx, actor, usecases.DescribeGridQuery{Limit: in.Limit, Offset: in.Offset})
		if err != nil {
			return DescribeGridOutput{}, err
		}
		return toDescribeGridOutput(res), nil
	})

	authx.Authed(api, auth, huma.Operation{
		OperationID: "apply" + tagTitle + "Grid",
		Method:      http.MethodPost,
		Path:        cfg.BasePath + "/grid",
		Summary:     "Apply bulk changes to " + summaryLower,
		Tags:        []string{cfg.Tag},
	}, func(ctx context.Context, actor *iam.Actor, in *ApplyGridInput) (ApplyGridOutput, error) {
		res, err := applyUC.Execute(ctx, actor, usecases.ApplyGridCommand{
			SchemaVersion: in.Body.SchemaVersion,
			Operations:    fromOpsDTO(in.Body.Operations),
		})
		if err != nil {
			return ApplyGridOutput{}, err
		}
		return toApplyGridOutput(res), nil
	})

	authx.Authed(api, auth, huma.Operation{
		OperationID: "lookup" + tagTitle + "Grid",
		Method:      http.MethodPost,
		Path:        cfg.BasePath + "/grid/lookup",
		Summary:     "Batch lookup/search for " + summaryLower + " references",
		Tags:        []string{cfg.Tag},
	}, func(ctx context.Context, actor *iam.Actor, in *LookupGridInput) (LookupGridOutput, error) {
		res, err := lookupUC.Execute(ctx, actor, usecases.LookupGridQuery{
			Queries: fromLookupQueryDTO(in.Body.Queries),
		})
		if err != nil {
			return LookupGridOutput{}, err
		}
		return toLookupGridOutput(res, in.Body.Queries), nil
	})
}

func title(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
