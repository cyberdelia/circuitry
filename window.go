package circuitry

import (
	"container/ring"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrBucketSize = errors.New("bucket duration and size must divide equally")

type bucket struct {
	failures      int64
	successes     int64
	shortcircuits int64
}

func (b *bucket) MarkShortCircuited() {
	atomic.AddInt64(&b.shortcircuits, 1)
}

func (b *bucket) MarkSuccess() {
	atomic.AddInt64(&b.successes, 1)
}

func (b *bucket) MarkFailure() {
	atomic.AddInt64(&b.failures, 1)
}

func (b *bucket) Total() int64 {
	return b.Failures() + b.Successes() + b.ShortCircuited()
}

func (b *bucket) Errors() int64 {
	return b.Failures() + b.ShortCircuited()
}

func (b *bucket) Successes() int64 {
	return atomic.LoadInt64(&b.successes)
}

func (b *bucket) Failures() int64 {
	return atomic.LoadInt64(&b.failures)
}

func (b *bucket) ShortCircuited() int64 {
	return atomic.LoadInt64(&b.shortcircuits)
}

func (b *bucket) Reset() {
	atomic.StoreInt64(&b.failures, 0)
	atomic.StoreInt64(&b.successes, 0)
	atomic.StoreInt64(&b.shortcircuits, 0)
}

type Window struct {
	size int
	ring *ring.Ring
	m    sync.RWMutex
}

// Create a new window of duration d containing n buckets.
func NewWindow(n int, d time.Duration) (*Window, error) {
	if int(d.Nanoseconds())%n != 0 {
		return nil, ErrBucketSize
	}
	w := &Window{
		size: n,
		ring: seed(ring.New(n)),
	}
	go w.tick(time.Duration(int(d) / n))
	return w, nil
}

func (w *Window) tick(d time.Duration) {
	for _ = range time.Tick(d) {
		w.m.Lock()
		w.ring = rollup(w.ring)
		w.m.Unlock()
	}
}

func rollup(r *ring.Ring) *ring.Ring {
	n := r.Next()
	return seed(n)
}

func seed(r *ring.Ring) *ring.Ring {
	if r.Value == nil {
		r.Value = new(bucket)
	} else {
		r.Value.(*bucket).Reset()
	}
	return r
}

// Record a short-circuit.
func (w *Window) markShortCircuited() {
	w.m.RLock()
	defer w.m.RUnlock()

	bucket := w.ring.Value.(*bucket)
	bucket.MarkShortCircuited()
}

// Record a success.
func (w *Window) markSuccess() {
	w.m.RLock()
	defer w.m.RUnlock()

	bucket := w.ring.Value.(*bucket)
	bucket.MarkSuccess()
}

// Record a failure.
func (w *Window) markFailure() {
	w.m.RLock()
	defer w.m.RUnlock()

	bucket := w.ring.Value.(*bucket)
	bucket.MarkFailure()
}

// Return the total number of access (failures, success and short-circuit).
func (w *Window) Total() (total int64) {
	w.m.RLock()
	defer w.m.RUnlock()

	w.ring.Do(func(i interface{}) {
		if b, ok := i.(*bucket); ok {
			total += b.Total()
		}
	})
	return total
}

// Return the percentage of errors (failures and short-circuit).
func (w *Window) Errors() int64 {
	w.m.RLock()
	defer w.m.RUnlock()

	var total, failures int64
	w.ring.Do(func(i interface{}) {
		if b, ok := i.(*bucket); ok {
			failures += b.Errors()
			total += b.Total()
		}
	})
	if total > 0 {
		return int64((float64(failures) / float64(total)) * 100)
	}
	return 0
}

// Return the total of short-circuit.
func (w *Window) ShortCircuited() (shortcircuit int64) {
	w.m.RLock()
	defer w.m.RUnlock()

	w.ring.Do(func(i interface{}) {
		if b, ok := i.(*bucket); ok {
			shortcircuit += b.ShortCircuited()
		}
	})
	return shortcircuit
}

// Return the total of successes.
func (w *Window) Successes() (successes int64) {
	w.m.RLock()
	defer w.m.RUnlock()

	w.ring.Do(func(i interface{}) {
		if b, ok := i.(*bucket); ok {
			successes += b.Successes()
		}
	})
	return successes
}

// Return the total of failures.
func (w *Window) Failures() (failures int64) {
	w.m.RLock()
	defer w.m.RUnlock()

	w.ring.Do(func(i interface{}) {
		if b, ok := i.(*bucket); ok {
			failures += b.Failures()
		}
	})
	return failures
}

// Reset window statistics.
func (w *Window) Reset() {
	w.m.Lock()
	defer w.m.Unlock()

	w.ring.Do(func(i interface{}) {
		if b, ok := i.(*bucket); ok {
			b.Reset()
		}
	})
}
