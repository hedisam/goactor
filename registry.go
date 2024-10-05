package goactor

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	// DefaultRegistrySize is the default initial size for the registry.
	DefaultRegistrySize = 1024

	// registrySizeEnvVar can be used to provide a custom value for the Registry.
	registrySizeEnvVar = "GOACTOR_PROCESSES_REGISTRY_SIZE"
)

// registry is used to name goactor processes.
type registry struct {
	nameToPID map[string]*PID
	pidToName map[string]string
	mu        sync.RWMutex
}

var processRegistry *registry

func init() {
	size := DefaultRegistrySize
	if sizeEnvVar := strings.TrimSpace(os.Getenv(registrySizeEnvVar)); sizeEnvVar != "" {
		s, err := strconv.Atoi(sizeEnvVar)
		if err != nil {
			size = s
		}
	}

	processRegistry = &registry{
		nameToPID: make(map[string]*PID, size),
		pidToName: make(map[string]string, size),
	}
}

// Register associates a PID with the given name.
func Register(name string, pid ProcessIdentifier) error {
	processRegistry.mu.Lock()
	defer processRegistry.mu.Unlock()

	regPID, ok := processRegistry.nameToPID[name]
	if ok {
		return fmt.Errorf("name already taken by <%s>", regPID.id)
	}
	regName, ok := processRegistry.pidToName[pid.PID().id]
	if ok {
		return fmt.Errorf("pid has already been given another name %q", regName)
	}

	processRegistry.nameToPID[name] = pid.PID()
	return nil
}

// Unregister disassociates a PID from the given name.
func Unregister(name string) {
	processRegistry.mu.Lock()
	defer processRegistry.mu.Unlock()

	pid, ok := processRegistry.nameToPID[name]
	if !ok {
		return
	}
	delete(processRegistry.nameToPID, name)
	delete(processRegistry.pidToName, pid.id)
}

// WhereIs returns the associated PID with the given name.
func WhereIs(name string) (*PID, bool) {
	processRegistry.mu.RLock()
	defer processRegistry.mu.RUnlock()

	pid, ok := processRegistry.nameToPID[name]
	return pid, ok
}

// namedPID is used to distinguish a NamedPID from a normal PID.
type namedPID interface {
	PID() *PID
	namedPID()
}

// NamedPID can be used to find and send message to an actor registered by a name
type NamedPID string

// PID returns the PID registered by the given name.
func (name NamedPID) PID() *PID {
	pid, _ := WhereIs(string(name))
	return pid
}

func (name NamedPID) namedPID() {}
