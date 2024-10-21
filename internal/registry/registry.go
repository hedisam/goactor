package registry

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/hedisam/goactor/internal/intprocess"
)

const (
	// DefaultSize is the default initial registry size.
	DefaultSize = 1024

	// SizeEnvVar can be used to provide a custom value for the reg.
	SizeEnvVar = "GOACTOR_PROCESSES_REGISTRY_SIZE"
)

var (
	// ErrSelfDisposed is returned when no registered PID is found for the running actor.
	ErrSelfDisposed = errors.New("self pid not found, probably disposed")
)

var reg *Registry

// Registry is used to name goactor processes.
type Registry struct {
	// used for registering actors with a name
	nameToPID map[string]intprocess.PID
	pidToName map[string]string
	nameMu    sync.RWMutex

	// used for linking processes and other operations limited to a running actor itself e.g. setting trap_exit
	gIDToPID map[string]intprocess.PID
	gIDMu    sync.RWMutex

	// used for remote actors
	refToPID map[string]intprocess.PID
	refMu    sync.RWMutex
}

func InitRegistry(size int) {
	reg = &Registry{
		nameToPID: make(map[string]intprocess.PID, size),
		pidToName: make(map[string]string, size),
		gIDToPID:  make(map[string]intprocess.PID, size),
		refToPID:  make(map[string]intprocess.PID, size),
	}
}

func GetRegistry() *Registry {
	return reg
}

func (r *Registry) RegisterSelf(pid intprocess.PID) {
	gid, err := goroutineID()
	if err != nil {
		panic(err)
	}
	r.gIDMu.Lock()
	r.gIDToPID[gid] = pid
	r.gIDMu.Unlock()

	r.refMu.Lock()
	r.refToPID[pid.Ref()] = pid
	r.refMu.Unlock()
}

func (r *Registry) UnregisterSelf() {
	gid, _ := goroutineID()
	r.gIDMu.Lock()
	pid := r.gIDToPID[gid]
	delete(r.gIDToPID, gid)
	r.gIDMu.Unlock()

	r.refMu.Lock()
	delete(r.refToPID, pid.Ref())
	r.refMu.Unlock()
}

func (r *Registry) PIDByRef(ref string) (intprocess.PID, bool) {
	r.refMu.RLock()
	defer r.refMu.RUnlock()

	pid, ok := r.refToPID[ref]
	return pid, ok
}

func (r *Registry) Self() (intprocess.PID, error) {
	gid, err := goroutineID()
	if err != nil {
		panic(err)
	}

	r.gIDMu.RLock()
	defer r.gIDMu.RUnlock()

	pid, ok := r.gIDToPID[gid]
	if !ok {
		return nil, ErrSelfDisposed
	}
	return pid, nil
}

func RegisterNamed(name string, pid intprocess.PID) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("cannot register empty name")
	}
	if pid == nil {
		return errors.New("cannot register nil pid")
	}

	reg.nameMu.Lock()
	defer reg.nameMu.Unlock()

	regPID, ok := reg.nameToPID[name]
	if ok {
		return fmt.Errorf("name already taken by %q", regPID.Ref())
	}
	regName, ok := reg.pidToName[pid.Ref()]
	if ok {
		return fmt.Errorf("pid has already been given another name %q", regName)
	}

	reg.nameToPID[name] = pid
	return nil
}

// UnregisterNamed disassociates a PID from the given name.
func UnregisterNamed(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	reg.nameMu.Lock()
	defer reg.nameMu.Unlock()

	pid, ok := reg.nameToPID[name]
	if !ok {
		return
	}
	delete(reg.nameToPID, name)
	delete(reg.pidToName, pid.Ref())
}

// WhereIs returns the associated PID with the given name.
func WhereIs(name string) (intprocess.PID, bool) {
	reg.nameMu.RLock()
	defer reg.nameMu.RUnlock()

	pid, ok := reg.nameToPID[name]
	return pid, ok
}

func goroutineID() (string, error) {
	var buf [32]byte
	n := runtime.Stack(buf[:], false)
	idx := bytes.IndexByte(buf[:n], '[')
	if idx <= 0 {
		return "", errors.New("could not extract actor's goroutine ID from stacktrace")
	}
	prefix := "goroutine "
	return string(buf[len(prefix) : idx-1]), nil
}
