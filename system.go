package goactor

import (
	p "github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/internal/relations"
	"github.com/hedisam/goactor/mailbox"
)

func buildActor(mailboxBuilder MailboxBuilderFunc) (*Actor, *PID) {
	if mailboxBuilder == nil {
		mailboxBuilder = DefaultQueueMailbox
	}
	m := mailboxBuilder()

	relationManager := relations.NewRelation()

	localPID := p.NewLocalPID(m, relationManager)

	actor := setupNewActor(m, localPID, relationManager)

	return actor, NewPID(localPID)
}

func spawn(fn ActorFunc, actor *Actor) {
	defer dispose(actor)
	fn(actor)
}

func DefaultQueueMailbox() Mailbox {
	return mailbox.NewQueueMailbox(
		mailbox.DefaultUserMailboxCap,
		mailbox.DefaultMailboxTimeout,
		mailbox.DefaultGoSchedulerInterval)
}

var DefaultChanMailbox = func() Mailbox {
	return mailbox.NewChanMailbox()
}

var DefaultRingBufferMailbox = func() Mailbox {
	return mailbox.NewRingBufferMailbox()
}
