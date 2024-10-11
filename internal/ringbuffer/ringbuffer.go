package ringbuffer

import "iter"

// RingBuffer represents a fixed-size circular buffer.
type RingBuffer[T any] struct {
	data []T
}

// New creates a new ring queue with a given capacity.
func New[T any](capacity uint) *RingBuffer[T] {
	return &RingBuffer[T]{
		data: make([]T, 0, capacity),
	}
}

// Put adds an element to the queue, overwriting the oldest element if the queue is full.
func (r *RingBuffer[T]) Put(v T) {
	if len(r.data) == cap(r.data) {
		if cap(r.data) == 0 {
			return
		}
		r.data = r.data[1:]
	}
	r.data = append(r.data, v)
}

// Pull removes and returns the element at the front of the queue.
func (r *RingBuffer[T]) Pull() (T, bool) {
	if len(r.data) == 0 {
		var zero T
		return zero, false
	}

	v := r.data[0]
	r.data = r.data[1:]
	return v, true
}

// Get returns the first element in the queue without removing it.
func (r *RingBuffer[T]) Get() (T, bool) {
	if len(r.data) == 0 {
		var zero T
		return zero, false
	}
	return r.data[0], true
}

// Size returns the number of enqueued data in the ring buffer.
func (r *RingBuffer[T]) Size() uint {
	return uint(len(r.data))
}

// Iter returns a iter.Seq2 to make the RingBuffer work with a for-range loop.
func (r *RingBuffer[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range r.data {
			if !yield(v) {
				return
			}
		}
	}
}
