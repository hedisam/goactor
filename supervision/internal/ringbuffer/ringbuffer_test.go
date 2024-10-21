package ringbuffer_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hedisam/goactor/supervision/internal/ringbuffer"
)

func TestRingBuffer(t *testing.T) {
	tests := map[string]struct {
		cap            uint
		toAdd          []any
		expectedItems  []any
		expectedPull   any
		expectedPullOk bool
		expectedGet    any
		expectedGetOk  bool
	}{
		"cap 0: add one item": {
			cap:            0,
			toAdd:          []any{"one"},
			expectedItems:  []any{},
			expectedPull:   nil,
			expectedPullOk: false,
			expectedGet:    nil,
			expectedGetOk:  false,
		},
		"cap 1: no items to add": {
			cap:            1,
			toAdd:          []any{},
			expectedItems:  []any{},
			expectedPull:   nil,
			expectedPullOk: false,
			expectedGet:    nil,
			expectedGetOk:  false,
		},
		"cap 1: add one item": {
			cap:            1,
			toAdd:          []any{"one"},
			expectedItems:  []any{"one"},
			expectedPull:   "one",
			expectedPullOk: true,
			expectedGet:    nil,
			expectedGetOk:  false,
		},
		"cap 1: add two items": {
			cap:            1,
			toAdd:          []any{"one", "two"},
			expectedItems:  []any{"two"},
			expectedPull:   "two",
			expectedPullOk: true,
			expectedGet:    nil,
			expectedGetOk:  false,
		},
		"cap 2: add three items": {
			cap:            2,
			toAdd:          []any{"one", "two", "three"},
			expectedItems:  []any{"two", "three"},
			expectedPull:   "two",
			expectedPullOk: true,
			expectedGet:    "three",
			expectedGetOk:  true,
		},
		"cap 4: add three items": {
			cap:            4,
			toAdd:          []any{"one", "three", "two"},
			expectedItems:  []any{"one", "three", "two"},
			expectedPull:   "one",
			expectedPullOk: true,
			expectedGet:    "three",
			expectedGetOk:  true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := ringbuffer.New[any](test.cap)
			for v := range slices.Values(test.toAdd) {
				r.Put(v)
			}
			actual := make([]any, 0)
			for v := range r.Iter() {
				actual = append(actual, v)
			}
			require.Equal(t, test.expectedItems, actual)
			assert.EqualValues(t, len(test.expectedItems), r.Size())

			pulled, ok := r.Pull()
			assert.Equal(t, test.expectedPullOk, ok)
			assert.Equal(t, test.expectedPull, pulled)

			got, ok := r.Get()
			assert.Equal(t, test.expectedGetOk, ok)
			assert.Equal(t, test.expectedGet, got)
		})
	}
}
