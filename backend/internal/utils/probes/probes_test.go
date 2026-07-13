package probes

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunChecks_EmptyList(t *testing.T) {
	res, err := runChecks(context.Background(), nil, time.Second)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil results for empty checks, got %+v", res)
	}
}

func TestRunChecks_AllPass(t *testing.T) {
	checks := []NamedCheck{
		{Name: "a", Check: func(context.Context) error { return nil }},
		{Name: "b", Check: nil},
	}
	res, err := runChecks(context.Background(), checks, time.Second)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("want 2 results, got %d", len(res))
	}
	for _, r := range res {
		if !r.OK {
			t.Fatalf("check %s must be OK, got %+v", r.Name, r)
		}
	}
}

func TestRunChecks_CapturesFirstError(t *testing.T) {
	boom := errors.New("db is down")
	wobble := errors.New("cache is flaky")

	checks := []NamedCheck{
		{Name: "db", Check: func(context.Context) error { return boom }},
		{Name: "cache", Check: func(context.Context) error { return wobble }},
		{Name: "queue", Check: func(context.Context) error { return nil }},
	}

	res, err := runChecks(context.Background(), checks, time.Second)
	if err == nil {
		t.Fatal("expected firstErr, got nil")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("firstErr must wrap boom, got %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("want 3 results, got %d", len(res))
	}
	if res[0].OK || res[1].OK {
		t.Fatalf("db/cache must report OK=false; got %+v", res[:2])
	}
	if !res[2].OK {
		t.Fatalf("queue must be OK; got %+v", res[2])
	}
}

func TestRunChecks_RespectsTimeout(t *testing.T) {
	slow := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			return nil
		}
	}
	checks := []NamedCheck{{Name: "slow", Check: slow}}

	start := time.Now()
	res, err := runChecks(context.Background(), checks, 50*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed > 400*time.Millisecond {
		t.Fatalf("timeout not respected, elapsed=%s", elapsed)
	}
	if len(res) != 1 || res[0].OK {
		t.Fatalf("slow check must be not OK, got %+v", res)
	}
}
