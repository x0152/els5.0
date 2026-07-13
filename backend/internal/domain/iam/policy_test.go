package iam_test

import (
	"strings"
	"testing"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
)

func TestPasswordPolicy_Validate(t *testing.T) {
	cases := []struct {
		name    string
		policy  iam.PasswordPolicy
		input   string
		wantErr bool
	}{
		{name: "default_min_short", policy: iam.PasswordPolicy{}, input: "1234567", wantErr: true},
		{name: "default_min_ok", policy: iam.PasswordPolicy{}, input: "12345678"},
		{name: "custom_min_short", policy: iam.PasswordPolicy{MinLength: 12}, input: "12345678901", wantErr: true},
		{name: "custom_min_ok", policy: iam.PasswordPolicy{MinLength: 12}, input: "123456789012"},
		{name: "default_max_too_long", policy: iam.PasswordPolicy{}, input: strings.Repeat("a", 129), wantErr: true},
		{name: "default_max_ok_128", policy: iam.PasswordPolicy{}, input: strings.Repeat("a", 128)},
		{name: "custom_max_above_hard_limit_capped", policy: iam.PasswordPolicy{MaxLength: 9999}, input: strings.Repeat("a", 1025), wantErr: true},
		{name: "custom_max_short", policy: iam.PasswordPolicy{MinLength: 8, MaxLength: 16}, input: strings.Repeat("a", 17), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Validate(tc.input)
			if tc.wantErr {
				test.ErrIs(t, err, shared.ErrValidation)
				return
			}
			test.NoErr(t, err)
		})
	}
}

func TestPasswordPolicy_Compare(t *testing.T) {
	p := iam.PasswordPolicy{MinLength: 8}

	if err := p.Compare("password1", "password1"); err != nil {
		t.Fatalf("expected no error for matching, got %v", err)
	}
	test.ErrIs(t, p.Compare("password1", "different"), shared.ErrValidation)
	test.ErrIs(t, p.Compare("short", "short"), shared.ErrValidation)
}
