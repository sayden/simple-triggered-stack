package simple_triggered_stack

import (
	"testing"
	"time"

	"github.com/apex/log/handlers/text"

	"os"
	"sync"

	"github.com/apex/log"
)

type groupedItem struct {
	type_       string
	flushAction GroupFlusher
	size        int
}

func (g *groupedItem) Group() string {
	return g.type_
}

func (g *groupedItem) SetGroup(t string) {
	g.type_ = t
}

func (g *groupedItem) Size() int64 {
	return int64(g.size)
}

func (g *groupedItem) FlushAction() GroupFlusher {
	return g.flushAction
}

type printToConsoleFlushAction struct {
	notifications int
	wg            sync.WaitGroup
	totals        map[string][]int64
	sync.RWMutex
}

func (p *printToConsoleFlushAction) Flush(gs []Grouped) {
	p.Lock()
	defer p.Unlock()

	if len(gs) == 0 {
		log.Error("Group was empty")
		return
	}

	log.WithFields(log.Fields{
		"group": gs[0].Group(),
		"type":  "printToConsoleFlushAction",
	}).Info("Flushing")

	for _, v := range gs {
		if p.totals[v.Group()] == nil || len(p.totals[v.Group()]) == 0 {
			p.totals[v.Group()] = make([]int64, 2)
		}

		p.totals[v.Group()][0]++
		p.totals[v.Group()][1] += v.Size()
	}

	p.notifications++
	p.wg.Done()
}

func Test_Push(t *testing.T) {
	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)

	//We'll flush all groups every 2 seconds, with or without content
	timeout := time.Second * 2

	//We flush every 3 items or when their sizes sum 10 bytes
	w := NewWindowedStack(&WindowedStackOptions{
		Timeout:      timeout,
		MaxStackSize: 3,
		MaxBytesSize: 10,
	})

	flushReceiver := &printToConsoleFlushAction{
		totals: make(map[string][]int64),
	}

	newItemFunc := func(t string, s int) Grouped {
		return &groupedItem{
			type_:       t,
			flushAction: flushReceiver,
			size:        s,
		}
	}

	t.Run("push 3 'hello' to trigger a flush", func(t *testing.T) {
		flushReceiver.wg.Add(1)

		go func() {
			err := w.Push(newItemFunc("hello", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		go func() {
			err := w.Push(newItemFunc("hello", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		go func() {
			err := w.Push(newItemFunc("hello", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()

		flushReceiver.wg.Wait()
		if flushReceiver.notifications != 1 {
			t.Fatalf("Notifications weren't 1, they were %d", flushReceiver.notifications)
		}
		flushReceiver.notifications = 0
	})

	t.Run("Now a single very big hello", func(t *testing.T) {
		flushReceiver.wg.Add(1)

		go func() {
			err := w.Push(newItemFunc("hello", 100))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		flushReceiver.wg.Wait()

		if flushReceiver.notifications != 1 {
			t.Fatalf("Notifications weren't 2, they were %d", flushReceiver.notifications)
		}
		flushReceiver.notifications = 0
	})

	t.Run("Now we'll force a timeout trigger by inserting 2 'hello' and 2 'world'", func(t *testing.T) {
		now := time.Now()

		flushReceiver.wg.Add(3)
		go func() {
			err := w.Push(newItemFunc("hello", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		go func() {
			err := w.Push(newItemFunc("hello", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		go func() {
			err := w.Push(newItemFunc("world", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		go func() {
			err := w.Push(newItemFunc("world", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()

		go func() {
			err := w.Push(newItemFunc("triple", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		go func() {
			err := w.Push(newItemFunc("triple", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()
		go func() {
			err := w.Push(newItemFunc("triple", 1))
			if err != nil {
				log.WithError(err).Error("Could not push")
			}
		}()

		flushReceiver.wg.Wait()
		elapsed := time.Since(now).Seconds()

		//At least 5 seconds must have passed and a single flush notification should be received
		if flushReceiver.notifications != 3 {
			t.Fatalf("Notifications weren't 3, they were %d", flushReceiver.notifications)
		}
		flushReceiver.notifications = 0

		if (elapsed + 0.5) <= timeout.Seconds() {
			t.Fatalf("Elapsed time wasn't 2, was %f", elapsed)
		}
	})

	w.Quit()

	time.Sleep(time.Second)
}
