// A circuit breaker
package circuitry

import (
	"sync"
	"time"
)

// CircuitBreaker represents a circuit breaker
type CircuitBreaker struct {
	FailCounter  int
	FailMax      int
	ResetTimeout time.Duration
	State        circuitState
	StateLock    *sync.Mutex
}

// Create a new circuit breaker with failMax failures and a resetTimeout timeout 
func Breaker(failMax int, resetTimeout time.Duration) *CircuitBreaker {
	b := new(CircuitBreaker)
	b.FailCounter = 0
	b.FailMax = failMax
	b.ResetTimeout = resetTimeout
	b.StateLock = new(sync.Mutex)
	b.State = &closedCircuit{b}
	return b
}

// Reports if the circuit is closed
func (b *CircuitBreaker) IsClosed() bool {
	return b.State.BeforeCall()
}

// Reports if the circuit is open
func (b *CircuitBreaker) IsOpen() bool {
	return !b.State.BeforeCall()
}

// Pass error to the to the circuit breaker
func (b *CircuitBreaker) Error(err error) {
	if err == nil {
		b.State.HandleSuccess()
	} else {
		b.State.HandleFailure()
	}
}

// Close the circuit
func (b *CircuitBreaker) Close() {
	b.StateLock.Lock()
	b.FailCounter = 0
	b.State = &closedCircuit{b}
	b.StateLock.Unlock()
}

// Open the circuit
func (b *CircuitBreaker) Open() {
	b.StateLock.Lock()
	b.State = &openCircuit{time.Now(), b}
	b.StateLock.Unlock()
}

// Half-open the circuit
func (b *CircuitBreaker) HalfOpen() {
	b.StateLock.Lock()
	b.State = &halfopenCircuit{b}
	b.StateLock.Unlock()
}
