package circuitry

import (
	"errors"
	"testing"
	"time"
)

type alwaysFail struct{}

func (c *alwaysFail) Name() string {
	return "always-fail"
}

func (c *alwaysFail) Run() (interface{}, error) {
	return nil, errors.New("fail")
}

func (c *alwaysFail) Fallback() interface{} {
	return "fallback"
}

type alwaysSucceed struct{}

func (c *alwaysSucceed) Name() string {
	return "always-succeed"
}

func (c *alwaysSucceed) Run() (interface{}, error) {
	return "success", nil
}

func (c *alwaysSucceed) Fallback() interface{} {
	return "fallback"
}

func TestAlwaysFail(t *testing.T) {
	b := NewBreaker(40, 0, time.Minute, window())
	v := Execute(&alwaysFail{}, b)
	if v != "fallback" {
		t.Error("didn't execute fallback")
	}
}

func TestAlwaysSucceed(t *testing.T) {
	b := NewBreaker(40, 0, time.Minute, window())
	v := Execute(&alwaysSucceed{}, b)
	if v != "success" {
		t.Error("the fallback was executed")
	}
}
