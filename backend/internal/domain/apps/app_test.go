package apps_test

import (
	"testing"

	"github.com/els/backend/internal/domain/apps"
)

func TestAccessPolicy_Allows(t *testing.T) {
	cases := []struct {
		name   string
		policy apps.AccessPolicy
		email  string
		roles  []string
		want   bool
	}{
		{name: "allow_all", policy: apps.AccessPolicy{AllowAll: true}, email: "a@b", roles: []string{"customer"}, want: true},
		{name: "role_match", policy: apps.AccessPolicy{AllowRoles: []string{"admin"}}, roles: []string{"admin"}, want: true},
		{name: "role_no_match", policy: apps.AccessPolicy{AllowRoles: []string{"admin"}}, roles: []string{"expert"}, want: false},
		{name: "email_match_case_insensitive", policy: apps.AccessPolicy{AllowEmails: []string{"USER@EXAMPLE.COM"}}, email: "user@example.com", want: true},
		{name: "email_match_with_whitespace", policy: apps.AccessPolicy{AllowEmails: []string{"  user@example.com  "}}, email: "user@example.com", want: true},
		{name: "email_no_match", policy: apps.AccessPolicy{AllowEmails: []string{"x@y"}}, email: "a@b", want: false},
		{name: "empty_policy", policy: apps.AccessPolicy{}, email: "a@b", roles: []string{"admin"}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.policy.Allows(tc.email, tc.roles)
			if got != tc.want {
				t.Fatalf("Allows(%q, %v) = %v, want %v", tc.email, tc.roles, got, tc.want)
			}
		})
	}
}

func TestCatalog_HasExpectedShape(t *testing.T) {
	if len(apps.Catalog) == 0 {
		t.Fatalf("expected non-empty catalog")
	}
	seen := map[string]bool{}
	for _, a := range apps.Catalog {
		if a.ID == "" || a.Name == "" || a.URI == "" {
			t.Fatalf("catalog entry missing id/name/uri: %+v", a)
		}
		if seen[a.ID] {
			t.Fatalf("duplicate app id: %s", a.ID)
		}
		seen[a.ID] = true
	}
}
