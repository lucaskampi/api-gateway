package resilience

import (
	"context"
	"sync"
	"time"

	"api-gateway/internal/domain/resilience"
)

type CircuitBreaker struct {
	mu              sync.RWMutex
	circuitBreakers map[string]*circuit
	attempts        int
	backoff         time.Duration
}

type circuit struct {
	failures    int
	successes   int
	state       string
	lastFailure time.Time
	mu          sync.Mutex
}

const (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half-open"
)

func NewCircuitBreaker(attempts int, backoff time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		circuitBreakers: make(map[string]*circuit),
		attempts:        attempts,
		backoff:         backoff,
	}
}

func (cb *CircuitBreaker) getCircuit(key string) *circuit {
	cb.mu.RLock()
	c, exists := cb.circuitBreakers[key]
	cb.mu.RUnlock()

	if exists {
		return c
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if c, exists = cb.circuitBreakers[key]; exists {
		return c
	}

	c = &circuit{
		state: StateClosed,
	}
	cb.circuitBreakers[key] = c
	return c
}

func (cb *CircuitBreaker) Execute(ctx context.Context, op string, fn func() error) error {
	circuit := cb.getCircuit(op)

	circuit.mu.Lock()
	defer circuit.mu.Unlock()

	if circuit.state == StateOpen {
		if time.Since(circuit.lastFailure) > cb.backoff*time.Duration(cb.attempts) {
			circuit.state = StateHalfOpen
		} else {
			return context.DeadlineExceeded
		}
	}

	err := fn()

	if err != nil {
		circuit.failures++
		circuit.lastFailure = time.Now()
		if circuit.failures >= cb.attempts {
			circuit.state = StateOpen
		}
	} else {
		circuit.successes++
		circuit.failures = 0
		if circuit.state == StateHalfOpen {
			circuit.state = StateClosed
		}
	}

	return err
}

func New(attempts int, backoff time.Duration) resilience.CircuitBreaker {
	return NewCircuitBreaker(attempts, backoff)
}
