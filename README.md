# Generic event emitter / listener

## Usage

```go
em := event.NewEmitter[int](context.Background())

listener := em.NewListener(event.On(func(n int) bool {
    return n >= 100
}))

for _, n := range []int{5, 4, 120, 100, 1, 4} {
	go em.Emit(n)
}

var numbers []int

for {
    select {
       case n, ok := <-listener.Channel():
           if ok {
             numbers = append(numbers, n)
           }		   
    }
}

// numbers received: 120, 100
```

Many listeners can be registered to the same emitter safely.