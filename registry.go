package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"sync"
)

type registry struct {
	actors map[string]intlpid.InternalPID
	sync.RWMutex
}

var reg *registry

func init() {
	reg = &registry{
		actors: make(map[string]intlpid.InternalPID),
	}
}

func Register(name string, pid *p.PID) {
	reg.Lock()
	defer reg.Unlock()
	reg.actors[name] = pid.InternalPID()
}

func Unregister(name string) {
	// acquire a read-lock to check if there's a pid registered with the given name
	reg.RLock()
	_, ok := reg.actors[name]
	if ok {
		// we have a registered internal_pid with that name.
		// unlock the read-only lock, and then acquire a write-lock
		reg.RUnlock()

		reg.Lock()
		delete(reg.actors, name)
		reg.Unlock()
		return
	}
	// if we're here it means there is no internal_pid registered with that name
	reg.RUnlock()
}

func WhereIs(name string) (*p.PID, bool) {
	reg.RLock()
	defer reg.RUnlock()

	intlPID, ok := reg.actors[name]
	return p.ToPID(intlPID), ok
}
