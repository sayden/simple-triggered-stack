# simple-triggered-stack

[![GoDoc](https://godoc.org/github.com/sayden/simple-triggered-stack?status.svg)](https://godoc.org/github.com/sayden/simple-triggered-stack)

A simple stack that stacks items until some limit is reached, then executes a callback to flush all items

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/sayden/simple-triggered-stack"
)


type enqueuableString struct {
	data string
}

func (e *enqueuableString) Data() interface{} {
	return e.data
}

func main() {
	msgN := 10

	//Channel that the stack will use to notify us that flushing has complete
	quit := make(chan struct{})

	stackConfig := simple_triggered_stack.Config{
		MaxStack:           3,
		MaxIngestionBuffer: 10,
	}

	stackCallback := func(stack []simple_triggered_stack.Stackable) {
		for _, v := range stack {
			fmt.Printf("%s ", v.Data())
		}

		fmt.Println()
	}

	q := simple_triggered_stack.NewStack(quit, stackConfig, stackCallback)

	for i := 0; i < msgN; i++ {
		time.Sleep(time.Millisecond * 100)
		q.IngestionCh() <- &enqueuableString{fmt.Sprintf("Hello %d", i)}
	}

	//We have finished pushing, close ingestion channel. Queue will notify of successful flushing by closing quit ch
	close(q.IngestionCh())

	<-quit
}
```