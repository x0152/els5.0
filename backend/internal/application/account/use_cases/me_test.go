package usecases_test

import (
	"context"
	"testing"

	usecases "github.com/els/backend/internal/application/account/use_cases"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestMe_ReturnsActor(t *testing.T) {
	actor := iamtest.Admin(t)
	uc := usecases.NewMeUseCase()

	res, err := uc.Execute(context.Background(), actor)

	test.NoErr(t, err)
	if res.Actor != actor {
		t.Errorf("expected same actor, got %v", res.Actor)
	}
}

func TestMe_NilActor(t *testing.T) {
	uc := usecases.NewMeUseCase()

	res, err := uc.Execute(context.Background(), nil)

	test.NoErr(t, err)
	if res.Actor != nil {
		t.Errorf("expected nil actor, got %v", res.Actor)
	}
}
