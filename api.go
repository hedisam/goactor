package goactor

import (
	"fmt"
)

func NewParentActor(mailboxBuilder MailboxFunc) (*Actor, func()) {
	actor := buildActor(mailboxBuilder)
	return actor, actor.dispose
}

func Spawn(fn ActorFunc, mailboxBuilder MailboxFunc) PID {
	actor := buildActor(mailboxBuilder)
	go spawn(fn, actor)
	return actor
}

func Send(pid PID , msg interface{}) error {
	err := pid.sendMessage(msg)
	if err != nil {
		return fmt.Errorf("couldn't send message: %w", err)
	}
	return nil
}