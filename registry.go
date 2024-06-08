package goactor

import (
	"fmt"
	"sync"
)

// DefaultRegistrySize is the default initial size for the registry.
const DefaultRegistrySize = 1024

type registry struct {
	nameToPID map[string]*PID
	pidToName map[string]string
	mu        sync.RWMutex
}

var procRegistry *registry

// InitRegistry initiates the process registry for naming actors.
func InitRegistry(size uint) {
	procRegistry = &registry{
		nameToPID: make(map[string]*PID, size),
		pidToName: make(map[string]string, size),
	}
}

// Register associates a PID with the given name.
func Register(name string, pid ProcessIdentifier) error {
	procRegistry.mu.Lock()
	defer procRegistry.mu.Unlock()

	regPID, ok := procRegistry.nameToPID[name]
	if ok {
		return fmt.Errorf("name already taken by <%s>", regPID.id)
	}
	regName, ok := procRegistry.pidToName[pid.PID().id]
	if ok {
		return fmt.Errorf("pid has already been given another name %q", regName)
	}

	procRegistry.nameToPID[name] = pid.PID()
	return nil
}

// Unregister disassociates a PID from the given name.
func Unregister(name string) {
	procRegistry.mu.Lock()
	defer procRegistry.mu.Unlock()

	pid, ok := procRegistry.nameToPID[name]
	if !ok {
		return
	}
	delete(procRegistry.nameToPID, name)
	delete(procRegistry.pidToName, pid.id)
}

// WhereIs returns the associated PID with the given name.
func WhereIs(name string) (*PID, bool) {
	procRegistry.mu.RLock()
	defer procRegistry.mu.RUnlock()

	pid, ok := procRegistry.nameToPID[name]
	return pid, ok
}
