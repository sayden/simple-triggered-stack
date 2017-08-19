package main

import (
	"fmt"
	"time"

	"github.com/sayden/simple-triggered-stack"
)

type stackableString struct {
	simple_triggered_stack.Stackable
	data string
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
			if s, ok := v.(*stackableString); ok {
				fmt.Printf("%s ", s.data)
			}
		}

		fmt.Println()
	}

	stack := simple_triggered_stack.NewStack(quit, stackConfig, stackCallback)

	for i := 0; i < msgN; i++ {
		time.Sleep(time.Millisecond * 100)
		stack.IngestionCh() <- &stackableString{data: fmt.Sprintf("Hello %d", i)}
	}

	//We have finished pushing, close ingestion channel. Queue will notify of successful flushing by closing quit ch
	close(stack.IngestionCh())

	<-quit
}
