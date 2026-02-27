package resilience

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 10*time.Millisecond)

	assert.NotNil(t, cb)
	assert.NotNil(t, cb.circuitBreakers)
	assert.Equal(t, 3, cb.attempts)
	assert.Equal(t, 10*time.Millisecond, cb.backoff)
}

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 10*time.Millisecond)

	for i := 0; i < 2; i++ {
		err := cb.Execute(nil, "test-op", func() error {
			return assert.AnError
		})
		assert.Equal(t, assert.AnError, err)
	}

	circuit := cb.getCircuit("test-op")
	assert.Equal(t, StateOpen, circuit.state)
}

func TestCircuitBreaker_OpenRejects(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)

	err := cb.Execute(nil, "test-op", func() error {
		return assert.AnError
	})
	assert.Equal(t, assert.AnError, err)

	err = cb.Execute(nil, "test-op", func() error {
		return nil
	})
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestCircuitBreaker_HalfOpenSuccess(t *testing.T) {
	cb := NewCircuitBreaker(1, 1*time.Millisecond)

	err := cb.Execute(nil, "test-op", func() error {
		return assert.AnError
	})
	assert.Equal(t, assert.AnError, err)

	time.Sleep(5 * time.Millisecond)

	err = cb.Execute(nil, "test-op", func() error {
		return nil
	})
	assert.NoError(t, err)

	circuit := cb.getCircuit("test-op")
	assert.Equal(t, StateClosed, circuit.state)
}

func TestCircuitBreaker_MultipleOperations(t *testing.T) {
	cb := NewCircuitBreaker(5, 10*time.Millisecond)

	ops := []string{"op1", "op2", "op3"}

	for _, op := range ops {
		err := cb.Execute(nil, op, func() error {
			return nil
		})
		assert.NoError(t, err)

		circuit := cb.getCircuit(op)
		assert.Equal(t, StateClosed, circuit.state)
		assert.Equal(t, 1, circuit.successes)
	}
}

func TestCircuitBreaker_GetCircuit(t *testing.T) {
	cb := NewCircuitBreaker(3, 10*time.Millisecond)

	c1 := cb.getCircuit("test")
	assert.NotNil(t, c1)
	assert.Equal(t, StateClosed, c1.state)

	c2 := cb.getCircuit("test")
	assert.Same(t, c1, c2)
}

var _ = time.Sleep
