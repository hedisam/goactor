package intlpid

import (
	"github.com/google/uuid"
	"sync"
)

type MockInternalPID struct {
	id string
	sync.RWMutex
}

func (pid *MockInternalPID) ID() string {
	pid.RLock()
	if pid.id != "" {
		id := pid.id
		pid.RUnlock()
		return id
	}
	pid.RUnlock()

	pid.Lock()
	defer pid.Unlock()

	if pid.id == "" {
		pid.id = uuid.New().String()
	}
	return pid.id
}

func (pid *MockInternalPID) IsSupervisor() bool {return false}

func (pid *MockInternalPID) sendMessage(_ interface{}) error {
	return nil
}

func (pid *MockInternalPID) sendSystemMessage(_ interface{}) error {
	return nil
}

func (pid *MockInternalPID) link(to InternalPID) error {
	return nil
}

func (pid *MockInternalPID) unlink(who InternalPID) error {
	return nil
}

func (pid *MockInternalPID) addMonitor(parent InternalPID) error {
	return nil
}

func (pid *MockInternalPID) remMonitor(parent InternalPID) error {
	return nil
}

func (pid *MockInternalPID) shutdown(_ interface{}) {

}
