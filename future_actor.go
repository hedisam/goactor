package goactor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/pid"
	"time"
)

type FutureActor struct {
	mailbox Mailbox
	self    *pid.PID
}

func setupNewFutureActor(mailbox Mailbox, self intlpid.InternalPID) *FutureActor {
	return &FutureActor{
		mailbox: mailbox,
		self:    pid.ToPID(self),
	}
}

func (a *FutureActor) Self() *pid.PID {
	return a.self
}

func (a *FutureActor) Receive(handler func(message interface{}) (loop bool)) {
	defer a.dispose()
	a.mailbox.Receive(handler, a.systemMessageHandler)
}

func (a *FutureActor) ReceiveWithTimeout(timeout time.Duration, handler func(message interface{}) (loop bool)) {
	defer a.dispose()
	a.mailbox.ReceiveWithTimeout(timeout, handler, a.systemMessageHandler)
}

func (a *FutureActor) systemMessageHandler(_ interface{}) (loop bool) {
	return true
}

func (a *FutureActor) dispose() {
	a.mailbox.Dispose()
}
