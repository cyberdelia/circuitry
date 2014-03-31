package circuitry

import (
	"errors"
	"testing"
	"time"
)

var ErrDummy = errors.New("dummy error")

func window() *Window {
	w, _ := NewWindow(10, 10*time.Second)
	return w
}

func TestTripCircuit(t *testing.T) {
	b := NewBreaker(40, 0, time.Minute, window())
	b.MarkSuccess()
	b.MarkSuccess()
	b.MarkSuccess()
	b.MarkSuccess()

	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}

	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()

	if b.Allow() {
		t.Error("should not allow requests")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}
}

func TestTripCircuitOnFailuresAboveThreshold(t *testing.T) {
	b := NewBreaker(40, 0, time.Minute, window())
	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}

	b.MarkSuccess()
	b.MarkSuccess()
	b.MarkFailure()
	b.MarkSuccess()
	b.MarkFailure()
	b.MarkFailure()
	b.MarkSuccess()
	b.MarkFailure()
	b.MarkFailure()

	if b.Allow() {
		t.Error("should not allow requests")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}
}

func TestCircuitDoesNotTripOnFailuresBelowThreshold(t *testing.T) {
	b := NewBreaker(40, 0, time.Minute, window())
	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}

	b.MarkSuccess()
	b.MarkSuccess()
	b.MarkFailure()
	b.MarkSuccess()
	b.MarkSuccess()
	b.MarkSuccess()
	b.MarkFailure()
	b.MarkFailure()

	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}
}

func TestSingleTestOnOpenCircuitAfterTimeWindow(t *testing.T) {
	b := NewBreaker(40, 0, 200*time.Millisecond, window())

	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()

	if b.Allow() {
		t.Error("should not allow requests")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}

	time.Sleep(250 * time.Millisecond)

	if !b.Allow() {
		t.Error("should allow one request")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}
	if b.Allow() {
		t.Error("should not allow requests")
	}
}

func TestCircuitClosedAfterSuccess(t *testing.T) {
	b := NewBreaker(40, 0, 200*time.Millisecond, window())

	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()
	b.MarkShortCircuited()

	if b.Allow() {
		t.Error("should not allow requests")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}

	time.Sleep(250 * time.Millisecond)

	if !b.Allow() {
		t.Error("should allow one request")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}
	if b.Allow() {
		t.Error("should not allow requests")
	}

	b.MarkSuccess()

	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}
}

func TestLowVolumeDoesNotTripCircuit(t *testing.T) {
	b := NewBreaker(40, 5, 200*time.Millisecond, window())

	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()

	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}
}

func TestCircuitForceOpen(t *testing.T) {
	b := NewBreaker(40, 0, 200*time.Millisecond, window())
	b.ForceOpen()

	if b.Allow() {
		t.Error("should not allow requests")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}

	time.Sleep(250 * time.Millisecond)
	b.MarkSuccess()

	if b.Allow() {
		t.Error("should not allow request")
	}
	if b.IsClosed() {
		t.Error("should be open")
	}

	b.Close()
	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}
}

func TestCircuitForceClose(t *testing.T) {
	b := NewBreaker(40, 0, 200*time.Millisecond, window())
	b.ForceClose()

	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}

	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()
	b.MarkFailure()

	if !b.Allow() {
		t.Error("should allow requests")
	}
	if b.IsOpen() {
		t.Error("should be closed")
	}
}
