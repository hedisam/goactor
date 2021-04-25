package intlpid

import (
	"github.com/google/uuid"
)

type MockInternalPID struct {
	id string
}

func NewMockInternalPID() *MockInternalPID {
	return &MockInternalPID{
		id:      uuid.New().String(),
	}
}

func (pid *MockInternalPID) ID() string {
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
