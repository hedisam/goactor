package process

import (
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegistry(t *testing.T) {
	internalPID := intlpid.NewMockInternalPID()
	pid := p.ToPID(internalPID)
	name := "my pid"

	Register(name, pid)

	registeredPID, ok := WhereIs(name)
	if !assert.True(t, ok) {return}
	if !assert.NotNil(t, registeredPID) {return}
	if !assert.Equal(t, pid, registeredPID) {return}

	newInternalPID := intlpid.NewMockInternalPID()
	newPID := p.ToPID(newInternalPID)

	Register(name, newPID)

	registeredPID, ok = WhereIs(name)
	if !assert.True(t, ok) {return}
	if !assert.NotNil(t, registeredPID) {return}
	if !assert.Equal(t, newPID, registeredPID) {return}
	if !assert.NotEqual(t, pid, registeredPID) {return}

	Unregister(name)

	registeredPID, ok = WhereIs(name)
	assert.False(t, ok)
	assert.Nil(t, registeredPID)

	// unregister a name which doesn't exist in the registry
	Unregister(name)
	// now again unregister the same name to make sure the locks are freed as expected
	Unregister(name)
}
