package simple_triggered_stack

import "github.com/apex/log"

type stackGroup struct {
	items       []Grouped
	totalSize   int64
	flushAction GroupFlusher
}

func (s *stackGroup) flush() {
	if len(s.items) == 0 {
		//Removing an usnused group
		return
	}
	s.flushAction.Flush(s.items)

	s.items = make([]Grouped, 0)

	return
}

func (s *stackGroup) bytesSize() int64 {
	return s.totalSize
}

func (s *stackGroup) add(g Grouped) {
	if g == nil {
		log.Warn("'Grouped' instance is nil")
		return
	}

	s.items = append(s.items, g)
	s.totalSize += g.Size()
}

func (s *stackGroup) length() int {
	return len(s.items)
}

func newStackGroup(f GroupFlusher) *stackGroup {
	return &stackGroup{
		items:       make([]Grouped, 0),
		flushAction: f,
	}
}
