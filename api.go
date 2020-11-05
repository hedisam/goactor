package goactor

import (
	"fmt"
)

func NewParentActor(mailboxBuilder MailboxBuilderFunc) (*Actor, func(*Actor)) {
	actor, _ := buildActor(mailboxBuilder)
	return actor, dispose
}

func Spawn(fn ActorFunc, mailboxBuilder MailboxBuilderFunc) *PID {
	actor, pid := buildActor(mailboxBuilder)
	go spawn(fn, actor)

	return pid
}

func Send(pid *PID, msg interface{}) error {
	err := pid.intlPID.SendMessage(msg)
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
