package goactor

import (
	"github.com/hedisam/goactor/internal/pid"
	"time"
)

type FutureActor struct {
	mailbox Mailbox
	self    pid.InternalPID
}

func setupNewFutureActor(mailbox Mailbox, pid pid.InternalPID) *FutureActor {
	return &FutureActor{
		mailbox: mailbox,
		self:    pid,
	}
}

func (a *FutureActor) Self() *PID {
	return NewPID(a.self)
}

func (a *FutureActor) Receive(handler func(message interface{}) (loop bool)) {
	defer a.dispose()
	a.mailbox.Receive(handler, a.systemMessageHandler)
}

func (a *FutureActor) ReceiveWithTimeout(timeout time.Duration, handler func(message interface{}) (loop bool)) {
	defer a.dispose()
	a.mailbox.ReceiveWithTimeout(timeout, handler, a.systemMessageHandler)
}

func (a *FutureActor) systemMessageHandler(_ interface{}) (passToUser bool) {
	return true
}

func (a *FutureActor) dispose() {
	a.mailbox.Dispose()
}
