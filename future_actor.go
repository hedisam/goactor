package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/pid"
	"time"
)

type FutureActor struct {
	mailbox Mailbox
	self    *pid.PID
	msgHandler MessageHandler
}

func newFutureActor(mailbox Mailbox, self intlpid.InternalPID) *FutureActor {
	return &FutureActor{
		mailbox: mailbox,
		self:    pid.ToPID(self),
	}
}

func (a *FutureActor) Self() *pid.PID {
	return a.self
}

func (a *FutureActor) Receive(handler MessageHandler) error {
	defer a.dispose()
	a.msgHandler = handler
	return a.mailbox.Receive(handler, a.systemMessageHandler)
}

func (a *FutureActor) ReceiveWithTimeout(timeout time.Duration, handler MessageHandler) error {
	defer a.dispose()
	a.msgHandler = handler
	return a.mailbox.ReceiveWithTimeout(timeout, handler, a.systemMessageHandler)
}

func (a *FutureActor) systemMessageHandler(sysMsg interface{}) (loop bool) {
	return a.msgHandler(sysMsg)
}

func (a *FutureActor) dispose() {
	a.mailbox.Dispose()
}
