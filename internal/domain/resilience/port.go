package resilience

import (
	"context"
)

type CircuitBreaker interface {
	Execute(ctx context.Context, op string, fn func() error) error
}
