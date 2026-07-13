package usecases_test

import (
	"context"
	"testing"

	usecases "github.com/els/backend/internal/application/account/use_cases"
	"github.com/els/backend/internal/domain/apps"
	iamdom "github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func sampleCatalog() []apps.App {
	return []apps.App{
		{ID: "public", Access: apps.AccessPolicy{AllowAll: true}},
		{ID: "admin_only", Access: apps.AccessPolicy{AllowRoles: []string{"admin"}}},
		{ID: "expert_only", Access: apps.AccessPolicy{AllowRoles: []string{"expert"}}},
		{ID: "vip_only", Access: apps.AccessPolicy{AllowEmails: []string{"admin@example.com"}}},
	}
}

func appIDs(items []apps.App) []string {
	out := make([]string, 0, len(items))
	for _, a := range items {
		out = append(out, a.ID)
	}
	return out
}

func TestListApps_FiltersByRoleAndEmail(t *testing.T) {
	cases := []struct {
		name  string
		actor *iamdom.Actor
		want  []string
	}{
		{name: "admin_sees_public_admin_and_email_match", actor: iamtest.Admin(t), want: []string{"public", "admin_only", "vip_only"}},
		{name: "expert_sees_public_and_expert", actor: iamtest.Expert(t, vo.NewID()), want: []string{"public", "expert_only"}},
		{name: "customer_sees_only_public", actor: iamtest.Customer(t, vo.NewID()), want: []string{"public"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := usecases.NewListAppsUseCase(sampleCatalog())

			res, err := uc.Execute(context.Background(), tc.actor)

			test.NoErr(t, err)
			gotIDs := appIDs(res.Apps)
			if len(gotIDs) != len(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, gotIDs)
			}
			for i, want := range tc.want {
				if gotIDs[i] != want {
					t.Errorf("at %d: expected %s, got %s", i, want, gotIDs[i])
				}
			}
		})
	}
}

func TestListApps_DefaultsToProductionCatalog(t *testing.T) {
	uc := usecases.NewListAppsUseCase(nil)
	res, err := uc.Execute(context.Background(), iamtest.Admin(t))

	test.NoErr(t, err)
	if len(res.Apps) == 0 {
		t.Errorf("expected non-empty catalog when nil passed in")
	}
}
