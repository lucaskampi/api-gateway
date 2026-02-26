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
	circuit := cb.getCircuit(key)

	circuit.mu.Lock()
	defer circuit.mu.Unlock()

	if circuit.state == circuitStateOpen {
		if time.Since(circuit.lastFailure) > cb.backoff*time.Duration(cb.attempts) {
			circuit.state = circuitStateHalfOpen
		} else {
			return fiber.ErrServiceUnavailable
		}
	}

	err := fn()

	if err != nil {
		circuit.failures++
		circuit.lastFailure = time.Now()
		if circuit.failures >= cb.attempts {
			circuit.state = circuitStateOpen
		}
	} else {
		circuit.successes++
		circuit.failures = 0
		if circuit.state == circuitStateHalfOpen {
			circuit.state = circuitStateClosed
		}
	}

	return err
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

		err := globalCircuitBreaker.Execute(upstreamURL, func() error {
			return c.Next()
		})

		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "service temporarily unavailable",
			})
		}

		return nil
	}
}
