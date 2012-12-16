# circuitry

circuitry is a circuit breaker implementation in Go.

## Installation

Download and install :

```
$ go get github.com/cyberdelia/circuitry
```

Add it to your code :

```go
import "github.com/cyberdelia/circuitry"
```

## Usage

```go
circuit := circuitry.Breaker(5, time.Minute)
if circuit.IsClosed() {
	err := DangerousStuff()
	circuit.Error(err) 
}
```

Dealing with panic :

```go
func Safe() {
	circuit := circuitry.Breaker(5, time.Minute)
	defer func() {
		if e := recover(); e != nil {
			circuit.Error(e.(error))
		}
	}()
	if circuit.IsClosed() {
		MightPanic()
		circuit.Error(nil)
	}
}
```