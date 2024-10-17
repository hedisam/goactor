package goactor

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
)

const (
	// defaultRegistrySize is the default initial size for the processRegistry.
	defaultRegistrySize = 1024

	// registrySizeEnvVar can be used to provide a custom value for the Registry.
	registrySizeEnvVar = "GOACTOR_PROCESSES_REGISTRY_SIZE"
)

var (
	// ErrSelfDisposed is returned when no registered PID is found for the running actor.
	ErrSelfDisposed = errors.New("self pid not found, probably disposed")
)

// processRegistry is used to name goactor processes.
type processRegistry struct {
	// used for registering actors with a name
	nameToPID map[string]*PID
	pidToName map[string]string
	nameMu    sync.RWMutex

	// used for linking processes and other operations limited to a running actor itself e.g. setting trap_exit
	gIDToPID map[string]*PID
	gIDMu    sync.RWMutex
}

var registry *processRegistry

func initRegistry() {
	var size int
	if sizeEnvVar := strings.TrimSpace(os.Getenv(registrySizeEnvVar)); sizeEnvVar != "" {
		s, err := strconv.Atoi(sizeEnvVar)
		if err != nil {
			logger.Warn("Could not convert registry size env var value to int, using the default value",
				"error", err,
				slog.String("env_var", registrySizeEnvVar),
				slog.Int("default_registry_size", defaultRegistrySize),
			)
			s = defaultRegistrySize
		}
		size = s
	}

	registry = &processRegistry{
		nameToPID: make(map[string]*PID, size),
		pidToName: make(map[string]string, size),
		gIDToPID:  make(map[string]*PID, size),
	}
}

func (r *processRegistry) registerSelf(pid *PID) {
	gid, err := goroutineID()
	if err != nil {
		panic(err)
	}
	r.gIDMu.Lock()
	r.gIDToPID[gid] = pid
	r.gIDMu.Unlock()
}

func (r *processRegistry) unregisterSelf() {
	gid, _ := goroutineID()
	r.gIDMu.Lock()
	delete(r.gIDToPID, gid)
	r.gIDMu.Unlock()
}

func (r *processRegistry) self() (*PID, error) {
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

// Register associates a PID with the given name.
func Register(name string, pid ProcessIdentifier) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("cannot register empty name")
	}
	if pid == nil {
		return errors.New("cannot register nil pid")
	}

	registry.nameMu.Lock()
	defer registry.nameMu.Unlock()

	regPID, ok := registry.nameToPID[name]
	if ok {
		return fmt.Errorf("name already taken by <%s>", regPID.id)
	}
	regName, ok := registry.pidToName[pid.PID().id]
	if ok {
		return fmt.Errorf("pid has already been given another name %q", regName)
	}

	registry.nameToPID[name] = pid.PID()
	return nil
}

// Unregister disassociates a PID from the given name.
func Unregister(names ...string) {
	if len(names) == 0 {
		return
	}

	registry.nameMu.Lock()
	defer registry.nameMu.Unlock()

	for name := range slices.Values(names) {
		unregister(name)
	}
}

func unregister(name string) {
	pid, ok := registry.nameToPID[name]
	if !ok {
		return
	}
	delete(registry.nameToPID, name)
	delete(registry.pidToName, pid.id)
}

// WhereIs returns the associated PID with the given name.
func WhereIs(name string) (*PID, bool) {
	registry.nameMu.RLock()
	defer registry.nameMu.RUnlock()

	pid, ok := registry.nameToPID[name]
	return pid, ok
}

// NamedPID can be used to find and send message to an actor registered by a name
type NamedPID string

// PID returns the PID registered by the given name.
func (name NamedPID) PID() *PID {
	pid, _ := WhereIs(string(name))
	return pid
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
