package ratelimit

import (
	"context"
	"sync"
	"time"

	"api-gateway/internal/domain/ratelimit"
)

type TokenBucket struct {
	tokens          map[string]*tokenBucket
	mu              sync.RWMutex
	rps             int
	burst           int
	refillMs        int
	cleanupInterval time.Duration
	maxAge          time.Duration
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

func NewTokenBucket(rps, burst int) *TokenBucket {
	tb := &TokenBucket{
		tokens:          make(map[string]*tokenBucket),
		rps:             rps,
		burst:           burst,
		refillMs:        1000,
		cleanupInterval: 5 * time.Minute,
		maxAge:          10 * time.Minute,
	}

	go tb.cleanup()
	return tb
}

func (tb *TokenBucket) cleanup() {
	ticker := time.NewTicker(tb.cleanupInterval)
	for range ticker.C {
		tb.mu.Lock()
		now := time.Now()
		for key, bucket := range tb.tokens {
			if now.Sub(bucket.lastRefill) > tb.maxAge {
				delete(tb.tokens, key)
			}
		}
		tb.mu.Unlock()
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	bucket, exists := tb.tokens[key]

	if !exists {
		tb.tokens[key] = &tokenBucket{
			tokens:     float64(tb.burst - 1),
			lastRefill: now,
		}
		return true, nil
	}

	elapsed := now.Sub(bucket.lastRefill).Milliseconds()
	refill := int64(elapsed) * int64(tb.rps) / 1000
	bucket.tokens += float64(refill)
	if bucket.tokens > float64(tb.burst) {
		bucket.tokens = float64(tb.burst)
	}
	bucket.lastRefill = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true, nil
	}

	return false, nil
}

func NewRateLimiter(rps, burst int) ratelimit.RateLimiter {
	return NewTokenBucket(rps, burst)
}
