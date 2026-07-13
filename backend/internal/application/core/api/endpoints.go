package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/core/use_cases"
	"github.com/els/backend/internal/domain/iam"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator    *authx.Authenticator
	IngestEvents     *usecases.IngestEventsUseCase
	MarkUnclear      *usecases.MarkUnclearUseCase
	ListEvents       *usecases.ListEventsUseCase
	ListCatalog      *usecases.ListCatalogUseCase
	ListDictionaries *usecases.ListDictionariesUseCase
	WipeData         *usecases.WipeDataUseCase
	DeleteRows       *usecases.DeleteRowsUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "ingestCoreEvents",
		Method:      http.MethodPost,
		Path:        "/api/v1/core/events",
		Summary:     "Ingest learning events",
		Tags:        []string{"core"},
	}, func(ctx context.Context, actor *iam.Actor, in *IngestInput) (IngestOutput, error) {
		cmd, err := toIngestEventsCommand(in)
		if err != nil {
			return IngestOutput{}, err
		}
		res, err := deps.IngestEvents.Execute(ctx, actor, cmd)
		if err != nil {
			return IngestOutput{}, err
		}
		return toIngestEventsOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "markCoreEventUnclear",
		Method:      http.MethodPost,
		Path:        "/api/v1/core/events/unclear",
		Summary:     "Mark a heard line as not understood",
		Tags:        []string{"core"},
	}, func(ctx context.Context, actor *iam.Actor, in *MarkUnclearInput) (MarkUnclearOutput, error) {
		res, err := deps.MarkUnclear.Execute(ctx, actor, toMarkUnclearCommand(in))
		if err != nil {
			return MarkUnclearOutput{}, err
		}
		return MarkUnclearOutput{Updated: res.Updated}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listCoreEvents",
		Method:      http.MethodGet,
		Path:        "/api/v1/core/events",
		Summary:     "List learning events by status",
		Tags:        []string{"core"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListInput) (ListOutput, error) {
		q, err := toListEventsQuery(in)
		if err != nil {
			return ListOutput{}, err
		}
		res, err := deps.ListEvents.Execute(ctx, actor, q)
		if err != nil {
			return ListOutput{}, err
		}
		return toListEventsOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listCoreCatalog",
		Method:      http.MethodGet,
		Path:        "/api/v1/core/catalog",
		Summary:     "List word and grammar catalog",
		Tags:        []string{"core"},
	}, func(ctx context.Context, actor *iam.Actor, _ *CatalogInput) (CatalogOutput, error) {
		res, err := deps.ListCatalog.Execute(ctx, actor)
		if err != nil {
			return CatalogOutput{}, err
		}
		return toCatalogOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listCoreDictionaries",
		Method:      http.MethodGet,
		Path:        "/api/v1/core/dictionaries",
		Summary:     "List column dictionaries",
		Tags:        []string{"core"},
	}, func(ctx context.Context, actor *iam.Actor, _ *DictionariesInput) (DictionariesOutput, error) {
		res, err := deps.ListDictionaries.Execute(ctx, actor)
		if err != nil {
			return DictionariesOutput{}, err
		}
		return toDictionariesOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "wipeCore",
		Method:      http.MethodDelete,
		Path:        "/api/v1/core/data",
		Summary:     "Wipe own events",
		Tags:        []string{"core"},
	}, func(ctx context.Context, actor *iam.Actor, _ *WipeInput) (WipeOutput, error) {
		if err := deps.WipeData.Execute(ctx, actor); err != nil {
			return WipeOutput{}, err
		}
		return WipeOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "deleteCoreRows",
		Method:      http.MethodDelete,
		Path:        "/api/v1/core/rows",
		Summary:     "Delete selected rows",
		Tags:        []string{"core"},
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteRowsInput) (DeleteRowsOutput, error) {
		n, err := deps.DeleteRows.Execute(ctx, actor, usecases.DeleteRowsCommand{Kind: in.Body.Kind, IDs: in.Body.IDs})
		if err != nil {
			return DeleteRowsOutput{}, err
		}
		return DeleteRowsOutput{Deleted: n}, nil
	})
}
