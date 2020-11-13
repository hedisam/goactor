package childstate

type ChildrenStateIterator struct {
	data   []*ChildState
	pos    int
	length int
}

func newChildrenStateIterator(data []*ChildState) *ChildrenStateIterator {
	return &ChildrenStateIterator{
		data:   data,
		pos:    0,
		length: len(data),
	}
}

func (i *ChildrenStateIterator) HasNext() bool {
	return i.length > 0 && i.pos < i.length
}

func (i *ChildrenStateIterator) Value() *ChildState {
	if !i.HasNext() {
		return nil
	}
	value := i.data[i.pos]
	i.pos++
	return value
}
