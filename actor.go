package goactor

import (
	"fmt"
	p "github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/internal/relations"
	"log"
	"sync/atomic"
	"time"
)

const (
	trapExitNo = iota
	trapExitYes
)

type Actor struct {
	relationManager relationManager
	mailbox         Mailbox
	trapExit        int32
	self            p.InternalPID
}

func setupNewActor(mailbox Mailbox, self p.InternalPID, manager relationManager) *Actor {
	a := &Actor{
		mailbox:         mailbox,
		relationManager: manager,
		trapExit:        trapExitNo,
		self:            self,
	}
	return a
}

func (a *Actor) TrapExit() bool {
	return atomic.LoadInt32(&a.trapExit) == trapExitYes
}

func (a *Actor) SetTrapExit(trap bool) {
	if trap {
		atomic.StoreInt32(&a.trapExit, trapExitYes)
		return
	}
	atomic.StoreInt32(&a.trapExit, trapExitNo)
}

func (a *Actor) Self() *PID {
	return NewPID(a.self)
}

func (a *Actor) Receive(handler func(message interface{}) (loop bool)) {
	a.mailbox.Receive(handler, a.systemMessageHandler)
}

func (a *Actor) ReceiveWithTimeout(timeout time.Duration, handler func(message interface{}) (loop bool)) {
	a.mailbox.ReceiveWithTimeout(timeout, handler, a.systemMessageHandler)
}

func (a *Actor) Link(pid *PID) error {
	// first we need to ask the other actor to link to this actor.
	err := pid.intlPID.Link(a.self)
	if err != nil {
		return fmt.Errorf("failed to add this actor to the target's linked actors list: %w", err)
	}
	// add the target actor to our linked actors list
	a.relationManager.AddLink(pid.intlPID)
	return nil
}

func (a *Actor) Unlink(pid *PID) error {
	// attempt to remove the link from the other actor
	err := pid.intlPID.Unlink(a.self)
	if err != nil {
		return fmt.Errorf("failed to remove this actor from the target's linked actors list: %w", err)
	}
	// remove the target actor from our linked actors list
	a.relationManager.RemoveLink(pid.intlPID)
	return nil
}

func (a *Actor) Monitor(pid *PID) error {
	// ask the child actor to be monitored by this actor.
	err := pid.intlPID.AddMonitor(a.self)
	if err != nil {
		return fmt.Errorf("failed to monitor: %w", err)
	}
	// save the child actor as monitored.
	a.relationManager.AddMonitored(pid.intlPID)
	return nil
}

func (a *Actor) Demonitor(pid *PID) error {
	// ask the target actor to be de-monitored by this actor.
	err := pid.intlPID.RemMonitor(a.self)
	if err != nil {
		return fmt.Errorf("failed to demonitor: %w", err)
	}
	// remove the child from our monitored actors list
	a.relationManager.RemoveMonitored(pid.intlPID)
	return nil
}

func (a *Actor) systemMessageHandler(sysMsg interface{}) (passToUser bool) {
	switch msg := sysMsg.(type) {
	case NormalExit:
		// some actor (linked or monitored) has exited with normal reason.
		relationType := a.relationManager.RelationType(msg.MsgFrom())
		if relationType == relations.MonitoredRelation || relationType == relations.LinkedRelation {
			// some child actor has exited normally. we should pass this message to the user.
			return true
		}
		return false
	case AbnormalExit:
		// some actor (linked or monitored) has exited with an abnormal reason.
		relationType := a.relationManager.RelationType(msg.MsgFrom())
		trapExit := atomic.LoadInt32(&a.trapExit)
		if relationType == relations.MonitoredRelation || (relationType == relations.LinkedRelation && trapExit == trapExitYes) {
			// if the message is from a monitored actor, or it's from a linked one but the current actor
			// is trapping exit messages, then we just need to pass the exit message to the user handler.
			return true
		} else if relationType == relations.LinkedRelation && trapExit == trapExitNo {
			// the terminated actor is linked and we're not trapping exit messages.
			// so we should panic with the same msg.
			panic(sysMsg)
		}

		// if the terminated actor is not linked, nor monitored, then we just ignore the system message.
		return false
	default:
		log.Printf("actor id: %v, unknown system message type: %v\n", a.Self().ID(), sysMsg)
		return false
	}
}

func dispose(a *Actor) {
	a.mailbox.Dispose()

	switch r := recover().(type) {
	case AbnormalExit:
		// the actor has received an exit message and called panic on it.
		// notifying linked and monitor actors.
		log.Printf("actor %v received an abnormal exit message from %v, reason: %v\n", a.Self().ID(), r.From.ID(), r.ExitReason())
		origin := r.MsgFrom()
		r.From = a.self
		a.notifyLinkedActors(r, origin)
		a.notifyMonitorActors(r)
	case NormalExit:
		// panic(NormalExit) has been called. so we just notify linked and monitor actors with a normal message.
		origin := r.MsgFrom()
		r.From = a.self
		a.notifyLinkedActors(r, origin)
		a.notifyMonitorActors(r)
	default:
		if r != nil {
			// something has gone wrong. notify with an AbnormalExit message.
			log.Printf("dispose: actor %v had a panic, reason: %v\n", a.Self().ID(), r)
			abnormal := AbnormalExit{
				From:   a.self,
				Reason: r,
			}
			// TODO: if this activity is a supervisor, then it hasn't had the chance to shutdown its children so we should do it now.
			a.notifyLinkedActors(abnormal, a.self)
			a.notifyMonitorActors(abnormal)
			return
		}
		// it's just a normal exit
		normal := NormalExit{From: a.self}
		a.notifyLinkedActors(normal, a.self)
		a.notifyMonitorActors(normal)
	}
}

func (a *Actor) notifyLinkedActors(msg ExitMessage, origin p.InternalPID) {
	iterator := a.relationManager.LinkedActors()
	for iterator.HasNext() {
		pid := iterator.Value()
		if origin.ID() != pid.ID() {
			if err := pid.SendSystemMessage(msg); err != nil {
				log.Printf("notifyLinkedActors: could not deliver exit message to pid: %v, origin: %v, err: %v\n",
					pid.ID(), msg.MsgFrom().ID(), err)
			}
		}
	}
}

func (a *Actor) notifyMonitorActors(msg ExitMessage) {
	iterator := a.relationManager.MonitorActors()
	for iterator.HasNext() {
		pid := iterator.Value()
		if err := pid.SendSystemMessage(msg); err != nil {
			log.Printf("notifyMonitorActors: could not deliver exit message to pid: %v, origin: %v, err: %v\n",
				pid.ID(), msg.MsgFrom().ID(), err)
		}
	}
}
