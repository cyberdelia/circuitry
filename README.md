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
w, _ := circuitry.NewWindow(10, 10*time.Second)
circuit := circuitry.NewBreaker(40, 4, time.Minute, w)
if circuit.Allow() {
	err := DangerousStuff()
	circuit.Error(err) 
}
```

Dealing with panic :

```go
func Safe() {
  w, _ := circuitry.NewWindow(10, 10*time.Second)
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
w, _ := circuitry.NewWindow(10, 10*time.Second)
circuit := circuitry.NewBreaker(40, 4, time.Minute, w)
if circuit.Allow() {
	if DangerousStuff() {
		circuit.MarkSuccess()
	} else {
		circuit.MarkFailure()
	}
}
```

Or via a Command :

```go
type Namer struct {
	Name string
}

func (n *Namer) Run() (interface{}, error) {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(4) >= 2 {
		return nil, fmt.Errorf("can't assign name: %s", n.Name)
	}
	return fmt.Sprintf("Your name is %s.", n.Name), nil
}

func (n *Namer) Fallback() interface{} {
	return fmt.Sprintf("Hello, %s.", n.Name)
}

func main() {
	cmd := &Namer{"Hal"}

	w, _ := circuitry.NewWindow(10, 10*time.Second)
	circuit := circuitry.NewBreaker(40, 4, time.Minute, w)
	value := circuitry.Execute(cmd, circuit)

	fmt.Println(value)
}
```