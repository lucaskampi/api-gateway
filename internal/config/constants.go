package config

import "time"

const (
	// Server timeouts (in milliseconds)
	DefaultReadTimeout  = 5000
	DefaultWriteTimeout = 5000
	DefaultIdleTimeout  = 30

	// Rate limiting defaults
	DefaultRateLimitRPS   = 100
	DefaultRateLimitBurst = 150
	RateLimitRefillMs     = 1000

	// Circuit breaker defaults
	DefaultCircuitBreakerAttempts  = 3
	DefaultCircuitBreakerBackoffMs = 100

	// Retry defaults
	DefaultRetryAttempts  = 3
	DefaultRetryBackoffMs = 100
	DefaultMaxBackoffSec  = 5

	// HTTP client defaults
	DefaultDialTimeout      = 5
	DefaultHTTPReadTimeout  = 10
	DefaultHTTPWriteTimeout = 10
	DefaultMaxIdleConns     = 100
)

var (
	RateLimitCleanupInterval = 5 * time.Minute
	RateLimitMaxAge          = 10 * time.Minute
)
