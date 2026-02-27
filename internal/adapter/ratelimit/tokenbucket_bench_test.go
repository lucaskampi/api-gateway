package ratelimit

import (
	"context"
	"testing"
)

func BenchmarkTokenBucket_Allow(b *testing.B) {
	tb := NewTokenBucket(100, 100)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tb.Allow(ctx, "bench-key")
	}
}

func BenchmarkTokenBucket_AllowUnderLimit(b *testing.B) {
	tb := NewTokenBucket(1000, 1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tb.Allow(ctx, "bench-key")
	}
}

func BenchmarkTokenBucket_MultipleKeys(b *testing.B) {
	tb := NewTokenBucket(100, 100)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key-" + string(rune(i%10+'0'))
		_, _ = tb.Allow(ctx, key)
	}
}
