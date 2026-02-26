package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	limiter := NewTokenBucket(10, 5)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "test-key")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("expected first 5 requests to be allowed")
		}
	}

	allowed, err := limiter.Allow(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected 6th request to be rate limited")
	}
}

func TestTokenBucket_DifferentKeys(t *testing.T) {
	limiter := NewTokenBucket(10, 1)

	ctx := context.Background()

	allowed1, err := limiter.Allow(ctx, "key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed1 {
		t.Error("expected first request to be allowed")
	}

	allowed2, err := limiter.Allow(ctx, "key2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !allowed2 {
		t.Error("expected request from different key to be allowed")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	limiter := NewTokenBucket(10, 2)

	ctx := context.Background()

	limiter.Allow(ctx, "test-key")
	limiter.Allow(ctx, "test-key")

	allowed, _ := limiter.Allow(ctx, "test-key")
	if allowed {
		t.Error("expected rate limit after burst")
	}

	time.Sleep(200 * time.Millisecond)

	allowed, _ = limiter.Allow(ctx, "test-key")
	if !allowed {
		t.Error("expected request to be allowed after refill")
	}
}
