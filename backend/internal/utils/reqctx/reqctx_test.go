package reqctx

import (
	"context"
	"testing"
)

func TestRequestID_RoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_abc")
	if got := RequestID(ctx); got != "req_abc" {
		t.Fatalf("want req_abc, got %q", got)
	}
}

func TestRequestID_Missing(t *testing.T) {
	if got := RequestID(context.Background()); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
	var nilCtx context.Context
	if got := RequestID(nilCtx); got != "" {
		t.Fatalf("nil ctx must be safe, got %q", got)
	}
}

func TestUser_RoundTrip(t *testing.T) {
	ctx := WithUser(context.Background(), User{ID: "u1", Email: "a@b", Role: "admin"})
	u, ok := UserOf(ctx)
	if !ok {
		t.Fatal("user not found in ctx")
	}
	if u.ID != "u1" || u.Email != "a@b" || u.Role != "admin" {
		t.Fatalf("unexpected user: %+v", u)
	}
}

func TestUser_Missing(t *testing.T) {
	if _, ok := UserOf(context.Background()); ok {
		t.Fatal("expected no user in ctx")
	}
	var nilCtx context.Context
	if _, ok := UserOf(nilCtx); ok {
		t.Fatal("nil ctx: expected no user")
	}
}
