/*
A circuit breaker

Circuit breaking with go errors:

	circuit := circuitry.NewBreaker(5, time.Minute)
	if circuit.IsClosed() {
		err := DangerousStuff()
		circuit.Error(err)
	}

*/
package circuitry

import (
	"sync"
	"time"
)

// CircuitBreaker represents a circuit breaker.
type CircuitBreaker struct {
	failures     counter
	failMax      uint64
	resetTimeout time.Duration
	state        circuitState
	m            sync.Mutex
}

// Create a new circuit breaker with failMax failures and a resetTimeout timeout.
func NewBreaker(failMax uint64, resetTimeout time.Duration) *CircuitBreaker {
	b := &CircuitBreaker{
		failMax:      failMax,
		resetTimeout: resetTimeout,
	}
	b.state = &closedCircuit{b}
	return b
}

// Reports if the circuit is closed.
func (b *CircuitBreaker) IsClosed() bool {
	return b.state.BeforeCall()
}

// Reports if the circuit is open.
func (b *CircuitBreaker) IsOpen() bool {
	return !b.state.BeforeCall()
}

// Pass error to the to the circuit breaker.
func (b *CircuitBreaker) Error(err error) {
	if err == nil {
		b.Success()
	} else {
		b.Failure()
	}
}

// Record a successful operation.
func (b *CircuitBreaker) Success() {
	b.state.HandleSuccess()
}

// Record a failure.
func (b *CircuitBreaker) Failure() {
	b.state.HandleFailure()
}

// Close the circuit.
func (b *CircuitBreaker) Close() {
	b.m.Lock()
	defer b.m.Unlock()
	b.failures.Reset()
	b.state = &closedCircuit{b}
}

// Open the circuit.
func (b *CircuitBreaker) Open() {
	b.m.Lock()
	defer b.m.Unlock()
	b.state = &openCircuit{time.Now(), b}
}

// Half-open the circuit.
func (b *CircuitBreaker) HalfOpen() {
	b.m.Lock()
	defer b.m.Unlock()
	b.state = &halfopenCircuit{b}
}
