package circuitry

import "time"

type circuitState interface {
	BeforeCall() bool
	HandleFailure()
	HandleSuccess()
}

type closedCircuit struct {
	Breaker *CircuitBreaker
}

func (c *closedCircuit) BeforeCall() bool {
	if c.Breaker.FailCounter >= c.Breaker.FailMax {
		c.Breaker.Open()
		return false
	}
	return true
}

func (c *closedCircuit) HandleFailure() {
	c.Breaker.FailCounter++
}

func (c *closedCircuit) HandleSuccess() {
	c.Breaker.FailCounter = 0
}

type openCircuit struct {
	OpenedAt time.Time
	Breaker  *CircuitBreaker
}

func (c *openCircuit) BeforeCall() (b bool) {
	if time.Now().Before(c.OpenedAt.Add(c.Breaker.ResetTimeout)) {
		b = false
	} else {
		c.Breaker.HalfOpen()
		b = true
	}
	return
}

func (c *openCircuit) HandleFailure() {}

func (c *openCircuit) HandleSuccess() {}

type halfopenCircuit struct {
	Breaker *CircuitBreaker
}

func (c *halfopenCircuit) BeforeCall() bool {
	return true
}

func (c *halfopenCircuit) HandleFailure() {
	c.Breaker.FailCounter++
	c.Breaker.Open()
}

func (c *halfopenCircuit) HandleSuccess() {
	c.Breaker.FailCounter = 0
	c.Breaker.Close()
}
