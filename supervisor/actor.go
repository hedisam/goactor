package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/models"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type Supervisor struct {
	relationManager models.RelationManager
	mailbox         models.Mailbox
	self            *p.PID
}

func newSupervisorActor(mailbox models.Mailbox, self intlpid.InternalPID, manager models.RelationManager) *Supervisor {
	sup := &Supervisor{
		mailbox:         mailbox,
		relationManager: manager,
		self:            p.ToPID(self),
	}
	return sup
}

func (sup *Supervisor) TrapExit() bool {
	return true
}

func (sup *Supervisor) Self() *p.PID {
	return sup.self
}

func (sup *Supervisor) Receive(handler func(message interface{}) (loop bool)) {
	sup.mailbox.Receive(handler, handler)
}

func (sup *Supervisor) Link(pid *p.PID) error {
	// first we need to ask the other actor to link to this actor.
	err := intlpid.Link(pid.InternalPID(), sup.self.InternalPID())
	if err != nil {
		return fmt.Errorf("failed to add this supervisor to the child's linked actors list: %w", err)
	}
	// add the target actor to our linked actors list
	sup.relationManager.AddLink(pid.InternalPID())
	return nil
}

func (sup *Supervisor) Unlink(pid *p.PID) error {
	// attempt to remove the link from the other actor
	err := intlpid.Unlink(pid.InternalPID(), sup.self.InternalPID())
	if err != nil {
		return fmt.Errorf("failed to remove this supervisor from the child's linked actors list: %w", err)
	}
	// remove the target actor from our linked actors list
	sup.relationManager.RemoveLink(pid.InternalPID())
	return nil
}

func (sup *Supervisor) systemMessageHandler(_ interface{}) (loop bool) {
	return true
}

func (sup *Supervisor) dispose() {
	sup.mailbox.Dispose()

	// todo: explicitly shutdown the children in case of uncontrolled abnormal exit (we need to access to the supervisor's service)

	var msg sysmsg.SystemMessage
	switch r := recover().(type) {
	case sysmsg.AbnormalExit:
		log.Println("[----] supervisor dispose: recovered from an AbnormalExit")
	case sysmsg.NormalExit:
		log.Println("[----] supervisor dispose: recovered from a NormalExit")
	case sysmsg.KillExit:
		log.Println("[----] supervisor dispose: recovered from a KillExit")
	case sysmsg.ShutdownCMD:
		log.Println("[----] supervisor dispose: recovered fro a ShutdownCMD")
	default:
		if r != nil {
			// something abnormal has happened.
			log.Printf("[----] supervisor dispose: %v had a panic, reason: %v\n", sup.self.ID(), r)
			msg = sysmsg.NewAbnormalExitMsg(sup.self.InternalPID(), r, nil)
		} else {
			// it's just a normal exit
			msg = sysmsg.NewNormalExitMsg(sup.self.InternalPID(), nil)
		}
	}

	sup.notifyRelatedActors(msg)
}

func (sup *Supervisor) notifyRelatedActors(msg sysmsg.SystemMessage) {
	linkedIterator := sup.relationManager.LinkedActors()
	for linkedIterator.HasNext() {
		sup.notify(linkedIterator.Value(), msg)
	}
	monitorIterator := sup.relationManager.MonitorActors()
	for monitorIterator.HasNext() {
		sup.notify(monitorIterator.Value(), msg)
	}
}

func (sup *Supervisor) notify(pid intlpid.InternalPID, msg sysmsg.SystemMessage) {
	if msg.Origin() != nil && msg.Origin().Sender() == pid {
		return
	}
	err := intlpid.SendSystemMessage(pid, msg)
	if err != nil {
		log.Printf("supervisor: notifyRelatedActors: could not deliver system message to pid: %v, sender: %v, err: %v\n",
			pid.ID(), msg.Sender().ID(), err)
	}
}
