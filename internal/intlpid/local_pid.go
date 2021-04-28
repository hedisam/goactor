package intlpid

import (
	"github.com/google/uuid"
)

type LocalPID struct {
	m            mailbox
	id           string
	relManager   relationManager
	isSupervisor bool
	shutdownFn   func()
}

func NewLocalPID(m mailbox, manager relationManager, isSupervisor bool, shutdown func()) *LocalPID {
	return &LocalPID{
		m:            m,
		id:           uuid.New().String(),
		relManager:   manager,
		isSupervisor: isSupervisor,
		shutdownFn:   shutdown,
	}
}

func (l *LocalPID) IsSupervisor() bool {
	return l.isSupervisor
}

func (l *LocalPID) shutdown(_ interface{}) {
	l.shutdownFn()
}

func (l *LocalPID) ID() string {
	return l.id
}

func (l *LocalPID) sendMessage(msg interface{}) error {
	return l.m.PushMessage(msg)
}

func (l *LocalPID) sendSystemMessage(msg interface{}) error {
	return l.m.PushSystemMessage(msg)
}

func (l *LocalPID) link(to InternalPID) error {
	return l.relManager.AddLink(to)
}

func (l *LocalPID) unlink(who InternalPID) error {
	return l.relManager.RemoveLink(who)
}

func (l *LocalPID) addMonitor(parent InternalPID) error {
	return l.relManager.AddMonitor(parent)
}

func (l *LocalPID) remMonitor(parent InternalPID) error {
	return l.relManager.RemoveMonitor(parent)
}
