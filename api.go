package goactor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
)

func NewParentActor(mailboxBuilder MailboxBuilderFunc) (*Actor, func(*Actor)) {
	actor, _ := buildActor(mailboxBuilder)
	return actor, dispose
}

func NewFutureActor() *FutureActor {
	return buildFutureActor()
}

func Spawn(fn ActorFunc, mailboxBuilder MailboxBuilderFunc) *p.PID {
	actor, pid := buildActor(mailboxBuilder)
	go spawn(fn, actor)

	return pid
}

func Send(pid *p.PID, msg interface{}) error {
	if pid.IsSupervisor() {
		return fmt.Errorf("can not send message to a supervisor, use supervisor ref instead")
	}
	err := intlpid.SendMessage(pid.InternalPID(), msg)
	if err != nil {
		return fmt.Errorf("couldn't send message: %w", err)
	}
	return nil
}

func SendNamed(name string, msg interface{}) error {
	pid, ok := WhereIs(name)
	if !ok {
		return fmt.Errorf("actor %s not found", name)
	}
	return Send(pid, msg)
}
