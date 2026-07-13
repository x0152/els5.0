package iam_test

import (
	"testing"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/utils/test"
)

func TestParseRole(t *testing.T) {
	cases := []struct {
		in   string
		want iam.Role
		ok   bool
	}{
		{in: "admin", want: iam.RoleAdmin, ok: true},
		{in: "expert", want: iam.RoleExpert, ok: true},
		{in: "customer", want: iam.RoleCustomer, ok: true},
		{in: "garbage", ok: false},
		{in: "", ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := iam.ParseRole(tc.in)
			if !tc.ok {
				if err == nil {
					t.Fatalf("expected error for %q", tc.in)
				}
				return
			}
			test.NoErr(t, err)
			if got != tc.want || got.String() != string(tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}
