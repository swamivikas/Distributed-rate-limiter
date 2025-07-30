package test

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"

	"rate-limiter/limiter"
)

func TestLimiterAllow(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	rdb, err := limiter.NewRedisClient(mr.Addr(), 0)
	if err != nil {
		t.Fatalf("redis client: %v", err)
	}

	l := limiter.New(rdb, 3, time.Second) // 3 tokens per second window
	key := "ratelimit:test"

	// First 3 allowed
	for i := 0; i < 3; i++ {
		ok, err := l.Allow(context.Background(), key)
		if err != nil || !ok {
			t.Fatalf("expected allowed at %d", i)
		}
	}
	// 4th should be blocked
	ok, _ := l.Allow(context.Background(), key)
	if ok {
		t.Fatalf("expected blocked")
	}

	// Wait for refill
	time.Sleep(time.Second)
	ok, _ = l.Allow(context.Background(), key)
	if !ok {
		t.Fatalf("expected allowed after refill")
	}
}
