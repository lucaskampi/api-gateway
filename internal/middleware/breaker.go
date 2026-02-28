package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
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
	circuitStateClosed   = "closed"
	circuitStateOpen     = "open"
	circuitStateHalfOpen = "half-open"
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
		state: circuitStateClosed,
	}
	cb.circuitBreakers[key] = c
	return c
}

func (cb *CircuitBreaker) Execute(key string, fn func() error) error {
	if !cb.Allow(key) {
		return fiber.ErrServiceUnavailable
	}

	err := fn()
	cb.Record(key, err == nil)
	return err
}

func (cb *CircuitBreaker) Allow(key string) bool {
	circuit := cb.getCircuit(key)

	circuit.mu.Lock()
	defer circuit.mu.Unlock()

	if circuit.state != circuitStateOpen {
		return true
	}

	if time.Since(circuit.lastFailure) > cb.backoff*time.Duration(cb.attempts) {
		circuit.state = circuitStateHalfOpen
		return true
	}

	return false
}

func (cb *CircuitBreaker) Record(key string, success bool) {
	circuit := cb.getCircuit(key)

	circuit.mu.Lock()
	defer circuit.mu.Unlock()

	if success {
		circuit.successes++
		circuit.failures = 0
		if circuit.state == circuitStateHalfOpen {
			circuit.state = circuitStateClosed
		}
		return
	}

	circuit.failures++
	circuit.lastFailure = time.Now()
	if circuit.failures >= cb.attempts {
		circuit.state = circuitStateOpen
	}
}

var globalCircuitBreaker *CircuitBreaker

func CircuitBreakerMiddleware(attempts int, backoff time.Duration) fiber.Handler {
	if globalCircuitBreaker == nil {
		globalCircuitBreaker = NewCircuitBreaker(attempts, backoff)
	}

	return func(c fiber.Ctx) error {
		upstream := c.Locals("upstream")
		if upstream == nil {
			return c.Next()
		}

		upstreamURL, ok := upstream.(string)
		if !ok {
			return c.Next()
		}

		if !globalCircuitBreaker.Allow(upstreamURL) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "service temporarily unavailable",
			})
		}

		err := c.Next()
		status := c.Response().StatusCode()
		success := err == nil && status < fiber.StatusInternalServerError
		globalCircuitBreaker.Record(upstreamURL, success)

		return err
	}
}
