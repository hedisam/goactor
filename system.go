package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	"github.com/hedisam/goactor/mailbox"
	p "github.com/hedisam/goactor/pid"
)

func buildActor(mailboxBuilder MailboxBuilderFunc) (*Actor, *p.PID) {
	if mailboxBuilder == nil {
		mailboxBuilder = DefaultQueueMailbox
	}
	m := mailboxBuilder()

	relationManager := relations.NewRelation()

	actor := setupNewActor(m, relationManager)

	localPID := intlpid.NewLocalPID(m, relationManager, false, actor.ctxCancel)
	pid := p.ToPID(localPID)
	actor.self = pid

	return actor, pid
}

func buildFutureActor() *FutureActor {
	noShutdown := func() {}
	m := mailbox.NewQueueMailbox(10, 10, mailbox.DefaultMailboxTimeout, mailbox.DefaultGoSchedulerInterval)
	localPID := intlpid.NewLocalPID(m, nil, false, noShutdown)
	featureActor := setupNewFutureActor(m, localPID)
	return featureActor
}

func spawn(fn ActorFunc, actor *Actor) {
	defer dispose(actor)
	fn(actor)
}

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
