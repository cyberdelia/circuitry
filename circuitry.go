// A circuit breaker
//
// Circuitry is a circuit breaker similar to Hystrix:
//
//     w, _ := NewWindow(10, 10*time.Second)
//     circuit := circuitry.NewBreaker(15, 100, time.Minute, w)
//     if circuit.Allow() {
//         err := DangerousStuff()
//         circuit.Error(err)
//     }
//
// Or via a Command:
//
//     type Namer struct {
//         Name string
//     }
//
//     func (n *Namer) Run() (interface{}, error) {
//         rand.Seed(time.Now().UnixNano())
//         if rand.Intn(4) >= 2 {
//             return nil, fmt.Errorf("can't assign name: %s", n.Name)
//         }
//         return fmt.Sprintf("Your name is %s.", n.Name), nil
//     }
//
//     func (n *Namer) Fallback() interface{} {
//         return fmt.Sprintf("Hello, %s.", n.Name)
//     }
//
//     func main() {
//         cmd := &Namer{"Hal"}
//
//         w, _ := circuitry.NewWindow(10, 10*time.Second)
//         circuit := circuitry.NewBreaker(40, 4, time.Minute, w)
//         value := circuitry.Execute(cmd, circuit)
//
//         fmt.Println(value)
//     }
//
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
	}
	if b.w.Total() >= b.volume && b.w.Errors() >= b.errors {
		b.Open()
		return false
	}
	return true
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
	b.w.markSuccess()
	if b.IsForced() {
		return
	}
	if b.IsOpen() {
		b.Close()
	}
}

// Record a failure.
func (b *CircuitBreaker) MarkFailure() {
	b.w.markFailure()
}

// Record a rejection.
func (b *CircuitBreaker) MarkShortCircuited() {
	b.w.markShortCircuited()
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
