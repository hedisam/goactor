package relations

import p "github.com/hedisam/goactor/internal/intlpid"

type RelationIterator struct {
	data   []p.InternalPID
	pos    int
	length int
}

func NewRelationIterator(data []p.InternalPID) *RelationIterator {
	return &RelationIterator{
		data:   data,
		pos:    0,
		length: len(data),
	}
}

func (i *RelationIterator) HasNext() bool {
	return i.length > 0 && i.pos < i.length
}

func (i *RelationIterator) Value() p.InternalPID {
	if !i.HasNext() {
		return nil
	}
	value := i.data[i.pos]
	i.pos++
	return value
}
