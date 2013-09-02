package circuitry

import (
	"sync/atomic"
)

type counter struct {
	c uint64
}

func (c *counter) Value() uint64 {
	return atomic.LoadUint64(&c.c)
}

// Increments the counter by 1.
func (c *counter) Increment() {
	atomic.AddUint64(&c.c, 1)
}

// Reset sets the counter to zero and returns its previous value.
func (c *counter) Reset() (n uint64) {
	for !atomic.CompareAndSwapUint64(&c.c, n, 0) {
		n = atomic.LoadUint64(&c.c)
	}
	return n
}
