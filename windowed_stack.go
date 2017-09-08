package simple_triggered_stack

import (
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/juju/errors"
)

type WindowedStackOptions struct {
	MaxBytesSize int64
	MaxStackSize int
	Timeout      time.Duration
}

type windowedStack struct {
	groups       map[string]*stackGroup
	flushCh      chan string
	maxBytesSize int64
	maxStackSize int
	timeout      time.Duration
	isSized      bool
	flushAllCh   chan struct{}
	quit         chan struct{}
	quitting     struct {
		inProgress bool
		mutex      sync.RWMutex
	}
	in chan Grouped
}

func (w *windowedStack) triggerGroupFlush(g string) {
	w.flushCh <- g
}

func (w *windowedStack) groupIsNew(g string) bool {
	return w.groups[g] == nil
}

func (w *windowedStack) recreateStackGroup(g Grouped) {
	fa := g.FlushAction()
	w.groups[g.Group()] = newStackGroup(fa)
}

func (w *windowedStack) Push(g Grouped) (err error) {
	log.WithField("group", g.Group()).Info("Pushing")

	if w.isQuitting() {
		err = errors.New("Channel actions are closed. Stack is flushing and closing...")
		return
	}

	if g == nil {
		log.Error("Grouped is 'nil'")
		return
	}

	w.in <- g

	return
}

func (w *windowedStack) groupBytesSize(g string) int64 {
	return w.groups[g].bytesSize()
}

func (w *windowedStack) add(g Grouped) {
	//Check if group already exists or create it if not
	if w.groupIsNew(g.Group()) {
		w.recreateStackGroup(g)
	}

	w.groups[g.Group()].add(g)

	//Trigger 'flush' if bytes size is over threshold and size based flushing is confifgured
	if w.isSized {
		if w.groupBytesSize(g.Group()) > w.maxBytesSize {
			go w.triggerGroupFlush(g.Group())
			return
		}
	}

	//Trigger flush if the stack has enough items
	if w.groups[g.Group()].length() >= w.maxStackSize {
		go w.triggerGroupFlush(g.Group())
	}
}

func (w *windowedStack) flushAll() {
	log.Debug("Flushing all")

	var wg sync.WaitGroup
	wg.Add(len(w.groups))

	for k := range w.groups {
		go func(g string) {
			w.groups[g].flush()
			wg.Done()
		}(k)
	}

	wg.Wait()
}

func (w *windowedStack) Quit() {
	log.Info("Quitting")

	w.quit <- struct{}{}
}

func (w *windowedStack) isQuitting() bool {
	w.quitting.mutex.RLock()
	defer w.quitting.mutex.RUnlock()

	return w.quitting.inProgress
}

func (w *windowedStack) launch() {
	//Launch time based flush
	go func() {
		for {
			<-time.After(w.timeout)
			if w.isQuitting() {
				return
			}

			w.flushAllCh <- struct{}{}
		}
	}()

	for {
		select {
		case v := <-w.in:
			//Stack
			w.add(v)
			log.Info("Added")
		case group := <-w.flushCh:
			log.WithField("group", group).Debug("Flushing")

			w.groups[group].flush()
		case <-w.flushAllCh:
			w.flushAll()
		case <-w.quit:
			//Graceful shutdown. Don't allow more pushes, Flush all, then return
			w.quitting.mutex.Lock()
			w.quitting.inProgress = true
			defer w.quitting.mutex.Unlock()

			w.flushAll()
			close(w.in)
			close(w.flushAllCh)
			close(w.flushCh)

			return
		}
	}
}

func NewWindowedStack(o ...*WindowedStackOptions) (w *windowedStack) {
	w = &windowedStack{
		groups:       make(map[string]*stackGroup),
		flushCh:      make(chan string, 0),
		flushAllCh:   make(chan struct{}, 0),
		quit:         make(chan struct{}, 0),
		in:           make(chan Grouped, 0),
		timeout:      time.Second * 3,
		maxBytesSize: 1024 * 1024 * 1,
		maxStackSize: 10,
	}

	if o != nil && len(o) > 0 {
		if o[0].MaxBytesSize != 0 {
			w.isSized = true
			w.maxBytesSize = o[0].MaxBytesSize
		}

		if o[0].MaxStackSize != 0 {
			w.maxStackSize = o[0].MaxStackSize
		}

		if o[0].Timeout != 0 {
			w.timeout = o[0].Timeout
		}
	}

	go w.launch()

	return
}
