package goactor

func buildActor(mailboxBuilder MailboxFunc) *Actor {
	mailboxFunc := DefaultMailbox
	if mailboxBuilder != nil {
		mailboxFunc = mailboxBuilder
	}

	mailbox := mailboxFunc()
	return setupNewActor(mailbox)
}

func spawn(fn ActorFunc, actor *Actor) {
	defer actor.dispose()
	fn(actor)
}

