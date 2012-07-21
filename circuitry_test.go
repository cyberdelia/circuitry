package circuitry

import (
	"github.com/bmizerany/assert"
	"testing"
	"time"
)

type DummyError string

func (e DummyError) Error() string {
	return string(e)
}

func TestSuccess(t *testing.T) {
	b := Breaker(5, 5e06)
	if b.IsClosed() {
		b.Error(nil)
	}
	assert.Equal(t, b.FailCounter, 0)
	assert.T(t, b.IsClosed())
}

func TestOneFail(t *testing.T) {
	b := Breaker(5, 5e06)
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	assert.Equal(t, b.FailCounter, 1)
	assert.T(t, b.IsClosed())
}

func TestSuccessAfterFail(t *testing.T) {
	b := Breaker(5, 5e06)
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	assert.Equal(t, b.FailCounter, 1)
	assert.T(t, b.IsClosed())
	if b.IsClosed() {
		b.Error(nil)
	}
	assert.Equal(t, b.FailCounter, 0)
	assert.T(t, b.IsClosed())
}

func TestSeveralFail(t *testing.T) {
	b := Breaker(3, 5e06)
	assert.Equal(t, b.FailMax, 3)
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}

	// Circuit should open
	assert.Equal(t, b.FailCounter, 3)
	assert.T(t, b.IsOpen())
}

func TestFailAfterTimeout(t *testing.T) {
	b := Breaker(3, 50e6)
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}

	// Circuit should be open
	assert.Equal(t, b.FailCounter, 3)
	assert.T(t, b.IsOpen())

	// Wait and check if circuit is half-open
	time.Sleep(50e6)
	assert.T(t, b.IsClosed())

	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}

	// Circuit should be open again 
	assert.Equal(t, b.FailCounter, 4)
	assert.T(t, b.IsOpen())
}

func TestSuccessAfterTimeout(t *testing.T) {
	b := Breaker(3, 5e06)
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}

	// Circuit should be open
	assert.Equal(t, b.FailCounter, 3)
	assert.T(t, b.IsOpen())

	// Wait and check if circuit is half-open
	time.Sleep(50e6)
	assert.T(t, b.IsClosed())

	if b.IsClosed() {
		b.Error(nil)
	}

	// Circuit should be closed again
	assert.Equal(t, b.FailCounter, 0)
	assert.T(t, b.IsClosed())
}

func TestFailureHalfOpen(t *testing.T) {
	b := Breaker(3, 5e06)
	b.HalfOpen()
	assert.Equal(t, b.FailCounter, 0)
	assert.T(t, b.IsClosed())
	if b.IsClosed() {
		b.Error(DummyError("dummy error"))
	}

	// Circuit should be open
	assert.Equal(t, b.FailCounter, 1)
	assert.T(t, b.IsOpen())
}

func TestSuccessHalfOpen(t *testing.T) {
	b := Breaker(3, 5e06)
	b.HalfOpen()
	assert.Equal(t, b.FailCounter, 0)
	assert.T(t, b.IsClosed())
	if b.IsClosed() {
		b.Error(nil)
	}

	// Circuit should be open
	assert.Equal(t, b.FailCounter, 0)
	assert.T(t, b.IsClosed())
}

func TestClose(t *testing.T) {
	b := Breaker(3, 5e06)
	b.Error(DummyError("dummy error"))
	b.Error(DummyError("dummy error"))
	b.Error(DummyError("dummy error"))

	// Circuit should be open
	assert.Equal(t, b.FailCounter, 3)
	assert.T(t, b.IsOpen())

	b.Close()

	// Circuit should be closed
	assert.Equal(t, b.FailCounter, 0)
	assert.T(t, b.IsClosed())
}

func TestRecovery(t *testing.T) {
	b := Breaker(1, 5e06)
	defer func() {
		if e := recover(); e != nil {
			b.Error(e.(error))
		}
		assert.Equal(t, b.FailCounter, 1)
		assert.T(t, b.IsOpen())
	}()
	panic(DummyError("dummy error"))
	b.Error(nil)
}
