package circuitry

import (
	"errors"
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

var ErrDummy = errors.New("dummy error")

func TestSuccess(t *testing.T) {
	b := NewBreaker(5, 5e06)
	if b.IsClosed() {
		b.Error(nil)
	}
	assert.T(t, b.IsClosed())
}

func TestOneFail(t *testing.T) {
	b := NewBreaker(5, 5e06)
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	assert.T(t, b.IsClosed())
}

func TestSuccessAfterFail(t *testing.T) {
	b := NewBreaker(5, 5e06)
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	assert.T(t, b.IsClosed())
	if b.IsClosed() {
		b.Error(nil)
	}
	assert.T(t, b.IsClosed())
}

func TestSeveralFail(t *testing.T) {
	b := NewBreaker(3, 5e06)
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	if b.IsClosed() {
		b.Error(ErrDummy)
	}

	// Circuit should open
	assert.T(t, b.IsOpen())
}

func TestFailAfterTimeout(t *testing.T) {
	b := NewBreaker(3, 50e6)
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	if b.IsClosed() {
		b.Error(ErrDummy)
	}

	// Circuit should be open
	assert.T(t, b.IsOpen())

	// Wait and check if circuit is half-open
	time.Sleep(50e6)
	assert.T(t, b.IsClosed())

	if b.IsClosed() {
		b.Error(ErrDummy)
	}

	// Circuit should be open again
	assert.T(t, b.IsOpen())
}

func TestSuccessAfterTimeout(t *testing.T) {
	b := NewBreaker(3, 5e06)
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	if b.IsClosed() {
		b.Error(ErrDummy)
	}
	if b.IsClosed() {
		b.Error(ErrDummy)
	}

	// Circuit should be open
	assert.T(t, b.IsOpen())

	// Wait and check if circuit is half-open
	time.Sleep(50e6)
	assert.T(t, b.IsClosed())

	if b.IsClosed() {
		b.Error(nil)
	}

	// Circuit should be closed again
	assert.T(t, b.IsClosed())
}

func TestFailureHalfOpen(t *testing.T) {
	b := NewBreaker(3, 5e06)
	b.HalfOpen()
	assert.T(t, b.IsClosed())
	if b.IsClosed() {
		b.Error(ErrDummy)
	}

	// Circuit should be open
	assert.T(t, b.IsOpen())
}

func TestSuccessHalfOpen(t *testing.T) {
	b := NewBreaker(3, 5e06)
	b.HalfOpen()
	assert.T(t, b.IsClosed())
	if b.IsClosed() {
		b.Error(nil)
	}

	// Circuit should be open
	assert.T(t, b.IsClosed())
}

func TestClose(t *testing.T) {
	b := NewBreaker(3, 5e06)
	b.Error(ErrDummy)
	b.Error(ErrDummy)
	b.Error(ErrDummy)

	// Circuit should be open
	assert.T(t, b.IsOpen())

	b.Close()

	// Circuit should be closed
	assert.T(t, b.IsClosed())
}

func TestRecovery(t *testing.T) {
	b := NewBreaker(1, 5e06)
	defer func() {
		if e := recover(); e != nil {
			b.Error(e.(error))
		}
		assert.T(t, b.IsOpen())
	}()
	panic(ErrDummy)
}
