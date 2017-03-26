package simple_triggered_stack

type Stackable interface {
	Data() interface{}
}

type stack struct {
	ingestion chan Stackable
	q         []Stackable
	curIndex  int
	cb        func([]Stackable)
}

type Config struct {
	MaxStack           int
	MaxIngestionBuffer int
}

func (s *stack) IngestionCh() chan Stackable {
	return s.ingestion
}

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
