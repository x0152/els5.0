package test

import (
	"errors"
	"testing"
)

func Must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func MustVal[T any](t *testing.T, v T, err error) T {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return v
}

func ErrIs(t *testing.T, err, target error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error wrapping %v, got nil", target)
	}
	if !errors.Is(err, target) {
		t.Fatalf("expected error wrapping %v, got %v", target, err)
	}
}

func NoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
