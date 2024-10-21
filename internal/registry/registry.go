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

var reg *ProcessRegistry

// ProcessRegistry is used to name goactor processes.
type ProcessRegistry struct {
	// used for registering actors with a name
	nameToPID map[string]intprocess.PID
	refToName map[string]string
	nameMu    sync.RWMutex

	// used for linking processes and other operations limited to a running actor itself e.g. setting trap_exit
	gIDToLocalProcess map[string]*intprocess.LocalProcess
	gIDMu             sync.RWMutex

	// used by the node server
	refToLocalProcess map[string]*intprocess.LocalProcess
	localProcessesRef sync.RWMutex
}

func InitRegistry(size int) {
	reg = &ProcessRegistry{
		nameToPID:         make(map[string]intprocess.PID, size),
		refToName:         make(map[string]string, size),
		gIDToLocalProcess: make(map[string]*intprocess.LocalProcess, size),
		refToLocalProcess: make(map[string]*intprocess.LocalProcess, size),
	}
}

func SelfRegistrar() *ProcessRegistry {
	return reg
}

func (r *ProcessRegistry) RegisterSelf(pid *intprocess.LocalProcess) {
	gid, err := goroutineID()
	if err != nil {
		panic(err)
	}
	r.gIDMu.Lock()
	r.gIDToLocalProcess[gid] = pid
	r.gIDMu.Unlock()

	r.localProcessesRef.Lock()
	r.refToLocalProcess[pid.Ref()] = pid
	r.localProcessesRef.Unlock()
}

func (r *ProcessRegistry) UnregisterSelf() {
	gid, _ := goroutineID()
	r.gIDMu.Lock()
	pid := r.gIDToLocalProcess[gid]
	delete(r.gIDToLocalProcess, gid)
	r.gIDMu.Unlock()

	r.localProcessesRef.Lock()
	delete(r.refToLocalProcess, pid.Ref())
	r.localProcessesRef.Unlock()
}

func LocalProcessByRef(ref string) (*intprocess.LocalProcess, bool) {
	reg.localProcessesRef.RLock()
	defer reg.localProcessesRef.RUnlock()

	pid, ok := reg.refToLocalProcess[ref]
	return pid, ok
}

func Self() (*intprocess.LocalProcess, error) {
	gid, err := goroutineID()
	if err != nil {
		panic(err)
	}

	reg.gIDMu.RLock()
	defer reg.gIDMu.RUnlock()

	pid, ok := reg.gIDToLocalProcess[gid]
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
	regName, ok := reg.refToName[pid.Ref()]
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
	delete(reg.refToName, pid.Ref())
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
