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
w, _ := NewWindow(10, 10*time.Second)
circuit := circuitry.NewBreaker(40, 4, time.Minute, w)
if circuit.Allow() {
	err := DangerousStuff()
	circuit.Error(err) 
}
```

Dealing with panic :

```go
func Safe() {
  w, _ := NewWindow(10, 10*time.Second)
	circuit := circuitry.NewBreaker(40, 4, time.Minute, w)
	defer func() {
		if e := recover(); e != nil {
			circuit.Error(e.(error))
		}
	}()
	if circuit.Allow() {
		MightPanic()
		circuit.Error(nil)
	}
}
```

Or if failure is not an error :

```go
w, _ := NewWindow(10, 10*time.Second)
circuit := circuitry.NewBreaker(40, 4, time.Minute, w)
if circuit.Allow() {
	if DangerousStuff() {
		circuit.MarkSuccess()
	} else {
		circuit.MarkFailure()
	}
}
```

