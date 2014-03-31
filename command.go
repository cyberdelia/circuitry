package circuitry

// Command represents a piece of code you want to run
// and a fallback value.
type Command interface {
	Run() (interface{}, error)
	Fallback() interface{}
}

// Execute the given command against the given circuit breaker.
func Execute(c Command, b *CircuitBreaker) interface{} {
	if b.Allow() {
		i, err := c.Run()
		if err != nil {
			b.MarkFailure()
			return c.Fallback()
		}
		b.MarkSuccess()
		return i
	}
	b.MarkShortCircuited()
	return c.Fallback()
}
