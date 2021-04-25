package pid

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToPID(t *testing.T) {
	internalPID := intlpid.NewMockInternalPID()

	pid := ToPID(internalPID)
	if !assert.NotNil(t, pid) {return}
	if !assert.Equal(t, internalPID, pid.InternalPID()) {return}
	assert.Equal(t, internalPID.ID(), pid.ID())
	assert.Equal(t, internalPID.IsSupervisor(), pid.IsSupervisor())

	pid2 := ToPID(internalPID)
	if !assert.NotNil(t, pid2) {return}
	assert.Equal(t, pid, pid2)

	internalPID2 := intlpid.NewMockInternalPID()
	pid3 := ToPID(internalPID2)
	assert.NotEqual(t, pid3, pid)

	nilPID := ToPID(nil)
	assert.Nil(t, nilPID)
}