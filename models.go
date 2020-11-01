package goactor

import (
	"github.com/hedisam/goactor/mailbox"
)

type Mailbox interface {
	Receive(msgHandler, sysMsgHandler func(interface{}) bool)
	PushMessage(msg interface{}) error
	PushSystemMessage(msg interface{}) error
	Dispose()
}

type ActorFunc func(actor *Actor)
type MailboxFunc func() Mailbox

var DefaultMailbox = func() Mailbox {
	return mailbox.NewChanMailbox()
}