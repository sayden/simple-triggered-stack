package main

import (
	"fmt"
	"time"

	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/kr/pretty"
	"github.com/sayden/simple-triggered-stack"
)

type stackableString struct {
	simple_triggered_stack.Stackable
	data string
}

type groupedItem struct {
	type_       string
	flushAction simple_triggered_stack.GroupFlusher
	iter        int
}

func (g *groupedItem) Group() string {
	return g.type_
}

func (g *groupedItem) SetGroup(t string) {
	g.type_ = t
}

func (g *groupedItem) Size() int64 {
	return 2
}

func (g *groupedItem) FlushAction() simple_triggered_stack.GroupFlusher {
	return g.flushAction
}

type printToConsoleFlushAction struct{}

func (p *printToConsoleFlushAction) Flush(gs []simple_triggered_stack.Grouped) error {
	for _, v := range gs {
		pretty.Println(v)
	}

	return nil
}

func (p *printToConsoleFlushAction) Notify(err error) {
	log.Info("Notified")
}

func main() {
	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)

	timeout := time.Second

	w := simple_triggered_stack.NewWindowedStack(&simple_triggered_stack.WindowedStackOptions{
		Timeout:      &timeout,
		MaxStackSize: 3,
		MaxBytesSize: 10,
	})

	i := -1
	newItemFunc := func() simple_triggered_stack.Grouped {
		i++
		return &groupedItem{
			type_:       "Hello",
			flushAction: &printToConsoleFlushAction{},
			iter:        i,
		}
	}

	w.Push(newItemFunc())
	w.Push(newItemFunc())
	w.Push(newItemFunc())

	time.Sleep(time.Second)

	w.Push(newItemFunc())
	w.Push(newItemFunc())

	w.Quit()
}

func ex1() {
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
