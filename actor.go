package goactor

import (
	"context"
	"fmt"
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/sysmsg"
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
	self            *p.PID
	ctx 			context.Context
	ctxCancel		func()
	msgHandler 		func(message interface{}) (loop bool)
}

func newActor(mailbox Mailbox, manager relationManager) *Actor {
	a := &Actor{
		mailbox:         mailbox,
		relationManager: manager,
		trapExit:        trapExitNo,
	}
	a.ctx, a.ctxCancel = context.WithCancel(context.Background())
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

func (a *Actor) Self() *p.PID {
	return a.self
}

func (a *Actor) Receive(handler func(message interface{}) (loop bool)) error {
	a.msgHandler = handler
	return a.mailbox.Receive(handler, a.systemMessageHandler)
}

func (a *Actor) ReceiveWithTimeout(timeout time.Duration, handler func(message interface{}) (loop bool)) error {
	a.msgHandler = handler
	return a.mailbox.ReceiveWithTimeout(timeout, handler, a.systemMessageHandler)
}

func (a *Actor) Link(pid *p.PID) error {
	if pid == nil {
		return ErrLinkNilTargetPID
	}
	// first we need to ask the other actor to link to this actor.
	err := intlpid.Link(pid.InternalPID(), a.self.InternalPID())
	if err != nil {
		return fmt.Errorf("failed to add this actor to the target's linked actors list: %w", err)
	}
	// add the target actor to our linked actors list
	a.relationManager.AddLink(pid.InternalPID())
	return nil
}

func (a *Actor) Unlink(pid *p.PID) error {
	if pid == nil {
		return ErrUnlinkNilTargetPID
	}
	// attempt to remove the link from the other actor
	err := intlpid.Unlink(pid.InternalPID(), a.self.InternalPID())
	if err != nil {
		return fmt.Errorf("failed to remove this actor from the target's linked actors list: %w", err)
	}
	// remove the target actor from our linked actors list
	a.relationManager.RemoveLink(pid.InternalPID())
	return nil
}

func (a *Actor) Monitor(pid *p.PID) error {
	if pid == nil {
		return ErrMonitorNilTargetPID
	}
	// ask the child actor to be monitored by this actor.
	err := intlpid.AddMonitor(pid.InternalPID(), a.self.InternalPID())
	if err != nil {
		return fmt.Errorf("failed to monitor: %w", err)
	}
	// save the child actor as monitored.
	a.relationManager.AddMonitored(pid.InternalPID())
	return nil
}

func (a *Actor) Demonitor(pid *p.PID) error {
	if pid == nil {
		return ErrDemonitorNilTargetPID
	}
	// ask the target actor to be de-monitored.
	err := intlpid.RemoveMonitor(pid.InternalPID(), a.self.InternalPID())
	if err != nil {
		return fmt.Errorf("failed to demonitor: %w", err)
	}
	// remove the child from our monitored actors list
	a.relationManager.RemoveMonitored(pid.InternalPID())
	return nil
}

func (a *Actor) shutdown() {
	a.ctxCancel()
}

func (a *Actor) Context() context.Context {
	return a.ctx
}

func (a *Actor) systemMessageHandler(sysMsg interface{}) (loop bool) {
	switch msg := sysMsg.(type) {
	case sysmsg.NormalExit:
		// some actor (linked or monitored) has exited with normal reason.
		relationType := a.relationManager.RelationType(msg.Sender())
		if relationType == relations.MonitoredRelation || relationType == relations.LinkedRelation {
			// some child actor has exited normally. we should pass this message to the user.
			return a.msgHandler(sysMsg)
		}
		break
	case sysmsg.AbnormalExit:
		// some actor (linked or monitored) has exited with an abnormal reason.
		relationType := a.relationManager.RelationType(msg.Sender())
		trapExit := atomic.LoadInt32(&a.trapExit)
		if relationType == relations.MonitoredRelation || (relationType == relations.LinkedRelation && trapExit == trapExitYes) {
			// if the message is from a monitored actor, or it's from a linked one but the current actor
			// is trapping exit messages, then we just need to pass the exit message to the user handler.
			return a.msgHandler(sysMsg)
		} else if relationType == relations.LinkedRelation && trapExit == trapExitNo {
			// the terminated actor is linked and we're not trapping exit messages.
			// so we should panic with the same msg.
			panic(sysMsg)
		}

		// if the terminated actor is not linked, nor monitored, then we just ignore the abnormal exit message.
		break
	case sysmsg.KillExit:
		// todo: implement
		break
	case sysmsg.ShutdownCMD:
		// todo: implement
		break
	default:
		log.Printf("actor id: %v, unknown system message type: %v\n", a.Self().ID(), sysMsg)
	}
	return true
}

func dispose(a *Actor) {
	a.mailbox.Dispose()

	var msg sysmsg.SystemMessage
	switch r := recover().(type) {
	case sysmsg.AbnormalExit:
		// the actor has received an exit message and called panic on it.
		// notifying linked and monitor actors.
		log.Printf("actor %v received an abnormal exit message from %v, reason: %v\n", a.Self().ID(), r.Sender().ID(), r.Reason())
		msg = sysmsg.NewAbnormalExitMsg(
			a.self.InternalPID(),
			"exiting by receiving an abnormal message",
			&r)
	case sysmsg.NormalExit:
		// panic(NormalExit) has been called. so we just notify linked and monitor actors with a normal message.
		msg = sysmsg.NewNormalExitMsg(a.self.InternalPID(), &r)
	case sysmsg.KillExit:
		msg = sysmsg.NewKillMessage(a.self.InternalPID(), "exiting by receiving a kill message", &r)
	case sysmsg.ShutdownCMD:
		msg = sysmsg.NewShutdownCMD(a.self.InternalPID(), "exiting by receiving a shutdown command", &r)
	default:
		if r != nil {
			// something has went wrong. notify with an AbnormalExit message.
			log.Printf("dispose: actor %v had a panic, reason: %v\n", a.Self().ID(), r)
			msg = sysmsg.NewAbnormalExitMsg(a.self.InternalPID(), r, nil)
		} else {
			// it's just a normal exit
			msg = sysmsg.NewNormalExitMsg(a.self.InternalPID(), nil)
		}
	}
	a.notifyRelatedActors(msg)
}

func (a *Actor) notifyRelatedActors(msg sysmsg.SystemMessage) {
	linkedIterator := a.relationManager.LinkedActors()
	for linkedIterator.HasNext() {
		a.notify(linkedIterator.Value(), msg)
	}
	monitorIterator := a.relationManager.MonitorActors()
	for monitorIterator.HasNext() {
		a.notify(monitorIterator.Value(), msg)
	}
}

func (a *Actor) notify(pid intlpid.InternalPID, msg sysmsg.SystemMessage) {
	if msg.Origin() != nil && msg.Origin().Sender() == pid {
		return
	}
	err := intlpid.SendSystemMessage(pid, msg)
	if err != nil {
		log.Printf("notifyRelatedActors: could not deliver system message to pid: %v, sender: %v, err: %v\n",
			pid.ID(), msg.Sender().ID(), err)
	}
}
