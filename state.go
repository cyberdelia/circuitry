package circuitry

import "time"

type circuitState interface {
	BeforeCall() bool
	HandleFailure()
	HandleSuccess()
}

type closedCircuit struct {
	breaker *CircuitBreaker
}

func (c *closedCircuit) BeforeCall() bool {
	if c.breaker.failures.Value() >= c.breaker.failMax {
		c.breaker.Open()
		return false
	}
	return true
}

func (c *closedCircuit) HandleFailure() {
	c.breaker.failures.Increment()
}

func (c *closedCircuit) HandleSuccess() {
	c.breaker.failures.Reset()
}

type openCircuit struct {
	openedAt time.Time
	breaker  *CircuitBreaker
}

func (c *openCircuit) BeforeCall() (b bool) {
	if time.Now().Before(c.openedAt.Add(c.breaker.resetTimeout)) {
		b = false
	} else {
		c.breaker.HalfOpen()
		b = true
	}
	return
}

func (c *openCircuit) HandleFailure() {}

func (c *openCircuit) HandleSuccess() {}

type halfopenCircuit struct {
	breaker *CircuitBreaker
}

func (c *halfopenCircuit) BeforeCall() bool {
	return true
}

func (c *halfopenCircuit) HandleFailure() {
	c.breaker.failures.Increment()
	c.breaker.Open()
}

func (c *halfopenCircuit) HandleSuccess() {
	c.breaker.failures.Reset()
	c.breaker.Close()
}
