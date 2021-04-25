package goactor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	"github.com/hedisam/goactor/mailbox"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/process"
)

func DefaultQueueMailbox() Mailbox {
	return mailbox.NewQueueMailbox(
		mailbox.DefaultUserMailboxCap,
		mailbox.DefaultSysMailboxCap,
		mailbox.DefaultMailboxTimeout,
		mailbox.DefaultGoSchedulerInterval)
}

var DefaultChanMailbox = func() Mailbox {
	return mailbox.NewChanMailbox(
		mailbox.DefaultUserMailboxCap,
		mailbox.DefaultSysMailboxCap,
		mailbox.DefaultMailboxTimeout)
}

func NewParentActor(mailboxBuilder MailboxBuilderFunc) (*Actor, func(*Actor)) {
	actor, _ := setupActor(mailboxBuilder)
	return actor, dispose
}

func NewFutureActor() *FutureActor {
	return setupFutureActor()
}

func Spawn(fn ActorFunc, mailboxBuilder MailboxBuilderFunc) *p.PID {
	actor, pid := setupActor(mailboxBuilder)
	go spawn(fn, actor)

	return pid
}

func Send(pid *p.PID, msg interface{}) error {
	if pid == nil {
		return ErrSendNilPID
	}
	if pid.IsSupervisor() {
		return ErrSendToSupervisor
	}
	err := intlpid.SendMessage(pid.InternalPID(), msg)
	if err != nil {
		return fmt.Errorf("send failed: %w", err)
	}
	return nil
}

func SendNamed(name string, msg interface{}) error {
	pid, ok := process.WhereIs(name)
	if !ok {
		return ErrSendNameNotFound
	}
	return Send(pid, msg)
}

func setupActor(mailboxBuilder MailboxBuilderFunc) (*Actor, *p.PID) {
	if mailboxBuilder == nil {
		mailboxBuilder = DefaultQueueMailbox
	}
	m := mailboxBuilder()

	relationManager := relations.NewRelation()

	actor := newActor(m, relationManager)

	localPID := intlpid.NewLocalPID(m, relationManager, false, actor.ctxCancel)
	pid := p.ToPID(localPID)
	actor.self = pid

	return actor, pid
}

func setupFutureActor() *FutureActor {
	noShutdown := func() {}
	m := mailbox.NewQueueMailbox(10, 10, mailbox.DefaultMailboxTimeout, mailbox.DefaultGoSchedulerInterval)
	localPID := intlpid.NewLocalPID(m, nil, false, noShutdown)
	featureActor := newFutureActor(m, localPID)
	return featureActor
}

func spawn(fn ActorFunc, actor *Actor) {
	defer dispose(actor)
	fn(actor)
}
