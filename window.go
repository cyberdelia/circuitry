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

type Window struct {
	size int
	ring *ring.Ring
	m    sync.RWMutex
}

func NewWindow(size int, d time.Duration) (*Window, error) {
	if int(d.Nanoseconds())%size != 0 {
		return nil, ErrBucketSize
	}
	w := &Window{
		size: size,
		ring: seed(ring.New(size)),
	}
	go w.tick(time.Duration(int(d) / size))
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
	r.Value = new(bucket)
	return r
}

func (w *Window) MarkShortCircuited() {
	w.m.RLock()
	defer w.m.RUnlock()

	bucket := w.ring.Value.(*bucket)
	bucket.MarkShortCircuited()
}

func (w *Window) MarkSuccess() {
	w.m.RLock()
	defer w.m.RUnlock()

	bucket := w.ring.Value.(*bucket)
	bucket.MarkSuccess()
}

func (w *Window) MarkFailure() {
	w.m.RLock()
	defer w.m.RUnlock()

	bucket := w.ring.Value.(*bucket)
	bucket.MarkFailure()
}

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

func (w *Window) ShortCircuited() (rejections int64) {
	w.m.RLock()
	defer w.m.RUnlock()

	w.ring.Do(func(i interface{}) {
		if b, ok := i.(*bucket); ok {
			rejections += b.ShortCircuited()
		}
	})
	return rejections
}

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

func (w *Window) Reset() {
	w.m.Lock()
	defer w.m.Unlock()

	w.ring = seed(ring.New(w.size))
}
