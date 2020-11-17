package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	"time"
)

type Mailbox interface {
	Receive(msgHandler, sysMsgHandler func(interface{}) bool)
	ReceiveWithTimeout(timeout time.Duration, msgHandler, sysMsgHandler func(interface{}) bool)
	PushMessage(msg interface{}) error
	PushSystemMessage(msg interface{}) error
	Dispose()
}

type relationManager interface {
	AddLink(pid intlpid.InternalPID)
	RemoveLink(pid intlpid.InternalPID)

	AddMonitored(pid intlpid.InternalPID)
	RemoveMonitored(pid intlpid.InternalPID)

	LinkedActors() *relations.RelationIterator
	MonitorActors() *relations.RelationIterator

	RelationType(pid intlpid.InternalPID) relations.RelationType
}

type ActorFunc func(actor *Actor)
type MailboxBuilderFunc func() Mailbox
