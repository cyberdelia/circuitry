/*
A circuit breaker

Circuit breaking with go errors:

	w, _ := NewWindow(10, 10*time.Second)
	circuit := circuitry.NewBreaker(15, 100, time.Minute, w)
	if circuit.Allow() {
		err := DangerousStuff()
		circuit.Error(err)
	}

*/
package circuitry

import (
	"sync"
	"sync/atomic"
	"time"
)

// CircuitBreaker represents a circuit breaker.
type CircuitBreaker struct {
	errors   int64
	volume   int64
	timeout  time.Duration
	open     int32
	openedAt int64
	forced   int32

	w *Window
	m sync.RWMutex
}

// Create a new circuit breaker.
func NewBreaker(errors, volume int64, timeout time.Duration, w *Window) *CircuitBreaker {
	return &CircuitBreaker{
		errors:  errors,
		volume:  volume,
		timeout: timeout,
		open:    btoi(false),
		w:       w,
	}
}

// Allow requests to proceed or not.
func (b *CircuitBreaker) Allow() bool {
	if b.IsForced() {
		return !b.IsOpen()
	}
	if b.IsOpen() {
		openedAt := atomic.LoadInt64(&b.openedAt)
		now := time.Now().UnixNano()
		if now > (openedAt + b.timeout.Nanoseconds()) {
			return atomic.CompareAndSwapInt64(&b.openedAt, openedAt, time.Now().UnixNano())
		}
		return false
	} else {
		if b.w.Total() >= b.volume && b.w.Errors() >= b.errors {
			b.Open()
			return false
		}
		return true
	}
}

// Reports if the circuit is closed.
func (b *CircuitBreaker) IsClosed() bool {
	return !b.IsOpen()
}

// Reports if the circuit is open.
func (b *CircuitBreaker) IsOpen() bool {
	return itob(atomic.LoadInt32(&b.open))
}

// Reports if the circuit is forced.
func (b *CircuitBreaker) IsForced() bool {
	return itob(atomic.LoadInt32(&b.forced))
}

// Pass error to the to the circuit breaker.
func (b *CircuitBreaker) Error(err error) {
	if err == nil {
		b.MarkSuccess()
	} else {
		b.MarkFailure()
	}
}

// Record a successful operation.
func (b *CircuitBreaker) MarkSuccess() {
	b.w.MarkSuccess()
	if b.IsForced() {
		return
	}
	if b.IsOpen() {
		b.Close()
	}
}

// Record a failure.
func (b *CircuitBreaker) MarkFailure() {
	b.w.MarkFailure()
}

// Record a rejection.
func (b *CircuitBreaker) MarkShortCircuited() {
	b.w.MarkShortCircuited()
}

// Close the circuit.
func (b *CircuitBreaker) Close() {
	atomic.StoreInt32(&b.forced, btoi(false))
	atomic.StoreInt32(&b.open, btoi(false))
}

// Force close the circuit.
func (b *CircuitBreaker) ForceClose() {
	b.Close()
	atomic.StoreInt32(&b.forced, btoi(true))
}

// Open the circuit.
func (b *CircuitBreaker) Open() {
	b.w.Reset()
	atomic.StoreInt32(&b.forced, btoi(false))
	atomic.StoreInt64(&b.openedAt, time.Now().UnixNano())
	atomic.StoreInt32(&b.open, btoi(true))
}

// Force open the circuit.
func (b *CircuitBreaker) ForceOpen() {
	b.Open()
	atomic.StoreInt32(&b.forced, btoi(true))
}

func btoi(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func itob(i int32) bool {
	if i == 1 {
		return true
	}
	return false
}
