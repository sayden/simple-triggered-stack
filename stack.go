package simple_triggered_stack

//Stackable must be satisfied for any interface that wants to be stacked using this library
type Stackable interface {
	Data() interface{}
}

type stack struct {
	ingestion chan Stackable
	q         []Stackable
	curIndex  int
	cb        func([]Stackable)
}

//Config used to configure the stack
type Config struct {
	//MaxStack represents the number of items that must be stacked before executing the callback
	MaxStack           int

	//MaxIngestionBuffer is the buzzer size of the channel that receives the data
	MaxIngestionBuffer int
}

//IngestionCh returns the channel the push items on to the stack
func (s *stack) IngestionCh() chan Stackable {
	return s.ingestion
}


//NewStack is the preferred way to create a new stack instance. quit channel will be 'notified' (by closing it) if the
//ingestion channel is closed AND the remaining data has been flushed using the callback
func NewStack(quit chan struct{}, c Config, cb func([]Stackable)) (s *stack) {
	s = &stack{
		q:         make([]Stackable, c.MaxStack),
		ingestion: make(chan Stackable, c.MaxIngestionBuffer),
		cb:        cb,
	}

	//Stacking mechanism
	go func() {
		for m := range s.ingestion {
			s.q[s.curIndex] = m
			s.curIndex++
			if s.curIndex == c.MaxStack {
				s.curIndex = 0
				s.cb(s.q)
			}
		}

		//Insert the last incoming data
		s.cb(s.q[:s.curIndex])

		close(quit)
	}()

	return
}
