package relations

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelationIterator(t *testing.T) {

	t.Run("zero items", func(t *testing.T) {
		it := NewRelationIterator([]intlpid.InternalPID{})

		assert.False(t, it.HasNext())
		assert.Nil(t, it.Value())
	})

	t.Run("one item", func(t *testing.T) {
		it := NewRelationIterator([]intlpid.InternalPID{getNewMockPID()})

		if !assert.True(t, it.HasNext()) {return}
		assert.NotNil(t, it.Value())
	})

	t.Run("multiple items", func(t *testing.T) {
		L := 10
		pids := make([]intlpid.InternalPID, L)
		for i := 0; i < L; i++ {
			pids[i] = getNewMockPID()
		}

		it := NewRelationIterator(pids)

		for i := 0; i < L; i++ {
			assert.True(t, it.HasNext())
			value := it.Value()
			assert.NotNil(t, value)
			assert.Equal(t, pids[i], value)
		}

		assert.False(t, it.HasNext())
		assert.Nil(t, it.Value())
	})

}
