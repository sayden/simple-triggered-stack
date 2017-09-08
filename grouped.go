package simple_triggered_stack

type Grouped interface {
	Group() string
	SetGroup(string)
	Size() int64
	FlushAction() GroupFlusher
}

type GroupFlusher interface {
	Flush([]Grouped)
}
