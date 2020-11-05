package goactor

import (
	p "github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/internal/relations"
)

func NewPID(intlPID p.InternalPID) *PID {
	return &PID{intlPID: intlPID}
}

type PID struct {
	intlPID p.InternalPID
}

func (p *PID) ID() string {
	return p.intlPID.ID()
}

type Mailbox interface {
	Receive(msgHandler, sysMsgHandler func(interface{}) bool)
	PushMessage(msg interface{}) error
	PushSystemMessage(msg interface{}) error
	Dispose()
}

type relationManager interface {
	AddLink(pid p.InternalPID)
	RemoveLink(pid p.InternalPID)

	AddMonitored(pid p.InternalPID)
	RemoveMonitored(pid p.InternalPID)

	LinkedActors() *relations.RelationIterator
	MonitorActors() *relations.RelationIterator

	RelationType(pid p.InternalPID) relations.RelationType
}

type relationIterator interface {
	HasNext() bool
	Value() p.InternalPID
}

type ActorFunc func(actor *Actor)
type MailboxBuilderFunc func() Mailbox
