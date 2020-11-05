package goactor

import (
	"github.com/hedisam/goactor/internal/pid"
	"sync"
)

type registry struct {
	actors map[string]pid.InternalPID
	sync.RWMutex
}

var reg *registry

func init() {
	reg = &registry{
		actors: make(map[string]pid.InternalPID),
	}
}

func Register(name string, pid *PID) {
	reg.Lock()
	defer reg.Unlock()
	reg.actors[name] = pid.intlPID
}

func Unregister(name string) {
	// acquire a read-lock to check if there's a PID registered with the given name
	reg.RLock()
	_, ok := reg.actors[name]
	if ok {
		// we have a registered pid with that name.
		// unlock the read-only lock, and then acquire a write-lock
		reg.RUnlock()

		reg.Lock()
		delete(reg.actors, name)
		reg.Unlock()
		return
	}
	// if we're here it means there is no pid registered with that name
	reg.RUnlock()
}

func WhereIs(name string) (*PID, bool) {
	reg.RLock()
	defer reg.RUnlock()

	intlPID, ok := reg.actors[name]
	return NewPID(intlPID), ok
}
