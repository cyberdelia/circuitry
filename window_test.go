package circuitry

import (
	"testing"
	"time"
)

func Testbucket(t *testing.T) {
	b := new(bucket)
	b.MarkShortCircuited()
	b.MarkFailure()
	b.MarkSuccess()

	if total := b.Total(); total != 3 {
		t.Fatalf("total should be 3 got %d", total)
	}

	if errors := b.Errors(); errors != 2 {
		t.Fatalf("errors should be 2 got %d", errors)
	}
}

func TestNewWindow(t *testing.T) {
	w, err := NewWindow(10, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if length := w.ring.Len(); length != 10 {
		t.Fatalf("wrong numbers of buckets got %d wants 10", length)
	}
	if b := w.ring.Value.(*bucket); b == nil {
		t.Fatal("first bucket is not initialized")
	}
}

func TestBadWindow(t *testing.T) {
	_, err := NewWindow(11, 10*time.Second)
	if err == nil {
		t.Fatal("should have thrown an error")
	}
}

func TestWindowAggregation(t *testing.T) {
	w, err := NewWindow(10, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	w.MarkShortCircuited()
	w.MarkFailure()
	w.MarkSuccess()

	if total := w.Total(); total != 3 {
		t.Fatalf("total should be 3 got %d", total)
	}

	if errors := w.Errors(); errors != 66 {
		t.Fatalf("errors should be 66 got %d", errors)
	}
}

func TestWindowRollout(t *testing.T) {
	w, err := NewWindow(4, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	w.MarkSuccess()
	time.Sleep(225 * time.Millisecond)

	if total := w.Total(); total != 0 {
		t.Fatalf("total should be zero, got %d", total)
	}
}
