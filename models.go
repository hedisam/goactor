package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	"time"
)

type Mailbox interface {
	Receive(msgHandler, sysMsgHandler func(interface{}) bool) error
	ReceiveWithTimeout(timeout time.Duration, msgHandler, sysMsgHandler func(interface{}) bool) error
	PushMessage(msg interface{}) error
	PushSystemMessage(msg interface{}) error
	Dispose()
}

type relationManager interface {
	AddLink(pid intlpid.InternalPID) error
	RemoveLink(pid intlpid.InternalPID) error

	AddMonitored(pid intlpid.InternalPID) error
	RemoveMonitored(pid intlpid.InternalPID) error

	LinkedActors() *relations.RelationIterator
	MonitorActors() *relations.RelationIterator

	RelationType(pid intlpid.InternalPID) relations.RelationType
	Dispose()
}

type ActorFunc func(actor *Actor)
type MailboxBuilderFunc func() Mailbox
type MessageHandler func(message interface{}) (loop bool)
