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
	state        circuitState
	lock         *sync.Mutex
}

// Create a new circuit breaker with failMax failures and a resetTimeout timeout 
func Breaker(failMax int, resetTimeout time.Duration) *CircuitBreaker {
	b := new(CircuitBreaker)
	b.FailCounter = 0
	b.FailMax = failMax
	b.ResetTimeout = resetTimeout
	b.lock = new(sync.Mutex)
	b.state = &closedCircuit{b}
	return b
}

// Reports if the circuit is closed
func (b *CircuitBreaker) IsClosed() bool {
	return b.state.BeforeCall()
}

// Reports if the circuit is open
func (b *CircuitBreaker) IsOpen() bool {
	return !b.state.BeforeCall()
}

// Pass error to the to the circuit breaker
func (b *CircuitBreaker) Error(err error) {
	if err == nil {
		b.state.HandleSuccess()
	} else {
		b.state.HandleFailure()
	}
}

// Close the circuit
func (b *CircuitBreaker) Close() {
	b.lock.Lock()
	b.FailCounter = 0
	b.state = &closedCircuit{b}
	b.lock.Unlock()
}

// Open the circuit
func (b *CircuitBreaker) Open() {
	b.lock.Lock()
	b.state = &openCircuit{time.Now(), b}
	b.lock.Unlock()
}

// Half-open the circuit
func (b *CircuitBreaker) HalfOpen() {
	b.lock.Lock()
	b.state = &halfopenCircuit{b}
	b.lock.Unlock()
}
