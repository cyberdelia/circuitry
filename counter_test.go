package circuitry

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestCounter(t *testing.T) {
	var c counter
	c.Increment()
	c.Increment()
	assert.Equal(t, c.c, uint64(2))
	n := c.Reset()
	assert.Equal(t, n, uint64(2))
	assert.Equal(t, c.c, uint64(0))
}
