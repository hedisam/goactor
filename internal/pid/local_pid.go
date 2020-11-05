package pid

import "github.com/google/uuid"

type LocalPID struct {
	m          mailbox
	id         string
	relManager relationManager
}

func NewLocalPID(m mailbox, manager relationManager) *LocalPID {
	return &LocalPID{
		m:          m,
		id:         uuid.New().String(),
		relManager: manager,
	}
}

func (l *LocalPID) ID() string {
	return l.id
}

func (l *LocalPID) SendMessage(msg interface{}) error {
	return l.m.PushMessage(msg)
}

func (l *LocalPID) SendSystemMessage(msg interface{}) error {
	return l.m.PushSystemMessage(msg)
}

func (l *LocalPID) Link(to InternalPID) error {
	l.relManager.AddLink(to)
	return nil
}

func (l *LocalPID) Unlink(who InternalPID) error {
	l.relManager.RemoveLink(who)
	return nil
}

func (l *LocalPID) AddMonitor(parent InternalPID) error {
	l.relManager.AddMonitor(parent)
	return nil
}

func (l *LocalPID) RemMonitor(parent InternalPID) error {
	l.relManager.RemoveMonitor(parent)
	return nil
}
