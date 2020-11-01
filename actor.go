package goactor

import (
	"github.com/google/uuid"
	"log"
	"sync/atomic"
)

const (
	trapExitNo = iota
	trapExitYes
	disposed
	notDisposed
)

type Actor struct {
	relations *Relations
	mailbox   Mailbox
	trapExit int32
	disposed int32
	msgHandler func(msg interface{}) (loop bool)
	id string
}

func setupNewActor(mailbox Mailbox) *Actor {
	a := &Actor{
		mailbox:   mailbox,
		relations: newRelations(),
		trapExit:  trapExitNo,
		disposed:  notDisposed,
		id: uuid.New().String(),
	}
	return a
}

func (a *Actor) TrapExit() bool {
	if atomic.LoadInt32(&a.trapExit) == trapExitYes {
		return true
	}
	return false
}

func (a *Actor) SetTrapExit(trap bool) {
	if trap {
		atomic.StoreInt32(&a.trapExit, trapExitYes)
		return
	}
	atomic.StoreInt32(&a.trapExit, trapExitNo)
}

func (a *Actor) ID() string {
	return a.id
}

func (a *Actor) Self() PID {
	return a
}

func (a *Actor) Receive(handler func(msg interface{}) (loop bool)) {
	a.msgHandler = handler
	a.mailbox.Receive(handler, a.systemMessageHandler)
}

func (a *Actor) Link(to PID) {
	a.relations.Link(to)
	to.getRelations().Link(a)
}

func (a *Actor) Unlink(from PID) {
	a.relations.Unlink(from)
	from.getRelations().Unlink(a)
}

func (a *Actor) Monitor(pid PID) {
	pid.getRelations().BeMonitored(a.Self())
}

func (a *Actor) Demonitor(pid PID) {
	pid.getRelations().BeDemonitored(a.Self())
}

func (a *Actor) getRelations() *Relations {
	return a.relations
}

func (a *Actor) sendMessage(msg interface{}) error {
	return a.mailbox.PushMessage(msg)
}

func (a *Actor) sendSystemMessage(msg interface{}) error {
	return a.mailbox.PushSystemMessage(msg)
}

func (a *Actor) systemMessageHandler(sysMsg interface{}) (loop bool) {
	switch msg :=  sysMsg.(type) {
	case NormalExit:
		// some actor (linked or monitored) has exited with normal reason.
		relationType := a.relations.RelationType(msg.MsgFrom())
		if relationType == MonitoredRelation || relationType == LinkedRelation {
			// some child actor has exited normally. we should pass this message to the user.
			return a.msgHandler(msg)
		}
		return true
	case AbnormalExit:
		// some actor (linked or monitored) has exited with an abnormal reason.
		relationType := a.relations.RelationType(msg.MsgFrom())
		trapExit := atomic.LoadInt32(&a.trapExit)

		if relationType == MonitoredRelation || (relationType == LinkedRelation && trapExit == trapExitYes) {
			// if the message is from a monitored actor, or it's from a linked one but the current actor
			// is trapping exit messages, then we just need to pass the exit message to the user handler.
			return a.msgHandler(msg)
		} else if relationType == LinkedRelation && trapExit == trapExitNo {
			// the terminated actor is linked and we're not trapping exit messages.
			// so we should panic with the same msg.
			panic(sysMsg)
		}

		// if the terminated actor is not linked, nor monitored, then we just ignore the system message.
		return true
	default:
		log.Printf("actor id: %v, unknown system message type: %v\n", a.Self().ID(), sysMsg)
		return true
	}
}

func (a *Actor) dispose() {
	// check if the actor has not already been disposed.
	if !atomic.CompareAndSwapInt32(&a.disposed, notDisposed, disposed) {
		return
	}

	a.mailbox.Dispose()

	err := recover()

	switch r := err.(type) {
	case AbnormalExit:
		// the actor has received an exit message and called panic on it.
		// notifying linked and monitor actors.
		log.Printf("actor %v received an abnormal exit message from %v, reason: %v\n", a.Self().ID(), r.From.ID(), r.ExitReason())
		origin := r.MsgFrom()
		r.From = a
		a.notifyLinkedActors(r, origin)
		a.notifyMonitorActors(r)
	case NormalExit:
		// panic(NormalExit) has been called. so we just notify linked and monitor actors with a normal message.
		origin := r.MsgFrom()
		r.From = a
		a.notifyLinkedActors(r, origin)
		a.notifyMonitorActors(r)
	default:
		if r != nil {
			// something has gone wrong. notify with an AbnormalExit message.
			log.Printf("actor %v had a panic, reason: %v\n", a.Self().ID(), r)
			abnormal := AbnormalExit{
				From: 		a,
				Reason:     r,
			}
			// TODO: if this activity is a supervisor, then it hasn't had the chance to shutdown its children so we should do it now.
			a.notifyLinkedActors(abnormal, a)
			a.notifyMonitorActors(abnormal)
			return
		}
		// it's just a normal exit
		normal := NormalExit{From: a}
		a.notifyLinkedActors(normal, a)
		a.notifyMonitorActors(normal)
	}
}

func (a *Actor) notifyLinkedActors(msg ExitMessage, origin PID) {
	for pid, _ := range a.relations.LinkedActors() {
		if origin.ID() != pid.ID() {
			if err := pid.sendSystemMessage(msg); err != nil {
				log.Printf("notifyLinkedActors: could not deliver exit message to pid: %v, origin: %v, err: %v\n",
					pid.ID(), msg.MsgFrom().ID(), err)
			}
		}
	}
}

func (a *Actor) notifyMonitorActors(msg ExitMessage) {
	for pid, _ := range a.relations.MonitorActors() {
		if err := pid.sendSystemMessage(msg); err != nil {
			log.Printf("notifyMonitorActors: could not deliver exit message to pid: %v, origin: %v, err: %v\n",
				pid.ID(), msg.MsgFrom().ID(), err)
		}
	}
}
