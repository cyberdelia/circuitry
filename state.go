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
	if c.breaker.FailCounter >= c.breaker.FailMax {
		c.breaker.Open()
		return false
	}
	return true
}

func (c *closedCircuit) HandleFailure() {
	c.breaker.FailCounter++
}

func (c *closedCircuit) HandleSuccess() {
	c.breaker.FailCounter = 0
}

type openCircuit struct {
	openedAt time.Time
	breaker  *CircuitBreaker
}

func (c *openCircuit) BeforeCall() (b bool) {
	if time.Now().Before(c.openedAt.Add(c.breaker.ResetTimeout)) {
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
	c.breaker.FailCounter++
	c.breaker.Open()
}

func (c *halfopenCircuit) HandleSuccess() {
	c.breaker.FailCounter = 0
	c.breaker.Close()
}
