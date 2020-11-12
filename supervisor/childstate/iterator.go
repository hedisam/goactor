package childstate

type childrenStateIterator struct {
	data   []*ChildState
	pos    int
	length int
}

func newChildrenStateIterator(data []*ChildState) *childrenStateIterator {
	return &childrenStateIterator{
		data:   data,
		pos:    0,
		length: len(data),
	}
}

func (i *childrenStateIterator) HasNext() bool {
	return i.length > 0 && i.pos < i.length
}

func (i *childrenStateIterator) Value() *ChildState {
	if !i.HasNext() {
		return nil
	}
	value := i.data[i.pos]
	i.pos++
	return value
}
