package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/models"
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

	r := recover()
	if r != nil {
		log.Println("[!] supervisor recovered from a panic")
	}
}
