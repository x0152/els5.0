package usecases_test

import (
	"context"
	"errors"
	"testing"

	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/utils/test"
)

func TestLogout_HappyPath(t *testing.T) {
	// arrange
	sess := &sessionsStub{}
	uc := usecases.NewLogoutUseCase(sess)

	// act
	err := uc.Execute(context.Background(), "tok")

	// assert
	test.NoErr(t, err)
	if len(sess.revokedTokens) != 1 || sess.revokedTokens[0] != "tok" {
		t.Fatalf("expected token revoked, got %+v", sess.revokedTokens)
	}
}

func TestLogout_RevokeFailurePropagates(t *testing.T) {
	// arrange
	sess := &sessionsStub{revokeErr: errors.New("db down")}
	uc := usecases.NewLogoutUseCase(sess)

	// act
	err := uc.Execute(context.Background(), "tok")

	// assert
	if err == nil {
		t.Fatalf("expected revoke error to propagate")
	}
}
