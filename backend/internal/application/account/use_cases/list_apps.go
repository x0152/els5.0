package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/apps"
	"github.com/els/backend/internal/domain/iam"
)

type ListAppsUseCase struct {
	catalog []apps.App
}

func NewListAppsUseCase(catalog []apps.App) *ListAppsUseCase {
	if catalog == nil {
		catalog = apps.Catalog
	}
	return &ListAppsUseCase{catalog: catalog}
}

type ListAppsResult struct {
	Apps []apps.App
}

func (uc *ListAppsUseCase) Execute(_ context.Context, actor *iam.Actor) (ListAppsResult, error) {
	// 1. Take the actor's email and role as input for the access check.
	email := actor.Account().Email().String()
	roles := []string{actor.Role().String()}

	// 2. Filter the catalog by each entry's ACL.
	out := make([]apps.App, 0, len(uc.catalog))
	for _, a := range uc.catalog {
		if a.Access.Allows(email, roles) {
			out = append(out, a)
		}
	}
	return ListAppsResult{Apps: out}, nil
}
